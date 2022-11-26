package dcs

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/safedep/gateway/services/gen"
	"github.com/safedep/gateway/services/pkg/common/config"
	"github.com/safedep/gateway/services/pkg/common/logger"
	"github.com/safedep/gateway/services/pkg/common/utils"
	"google.golang.org/protobuf/proto"
)

const (
	pdpEventIndexerName  = "PDP Events to OpenSearch Indexer"
	pdpEventIndexerGroup = "pdp-event-indexer-group"
)

type opensearchIndexer struct {
	client       *opensearch.Client
	shardedNames sync.Map
}

func opensearchPdpEventIndexerSubscription() eventSubscription[*gen.PolicyEvaluationEvent] {
	h := opensearchIndexer{}

	client, err := h.buildOpenSearchClient()
	if err != nil {
		panic(err)
	}

	h.client = client
	return h.pdpEventSubscription()
}

func (s *opensearchIndexer) pdpEventSubscription() eventSubscription[*gen.PolicyEvaluationEvent] {
	err := s.initOpenSearchIndex(config.DcsServiceConfig().GetPdpEventIndexName())
	if err != nil {
		panic(err)
	}

	cfg := config.PdpServiceConfig()
	return eventSubscription[*gen.PolicyEvaluationEvent]{
		name:  pdpEventIndexerName,
		group: pdpEventIndexerGroup,
		topic: cfg.GetPublisherConfig().GetTopicNames().GetPolicyAudit(),
		decoder: func(b []byte) (*gen.PolicyEvaluationEvent, error) {
			var event gen.PolicyEvaluationEvent
			err := proto.Unmarshal(b, &event)
			return &event, err
		},
		handler: s.pdpEventHandler(),
	}
}

func (s *opensearchIndexer) pdpEventHandler() eventSubscriptionHandler[*gen.PolicyEvaluationEvent] {
	return func(event *gen.PolicyEvaluationEvent) error {
		return s.handle(event)
	}
}

func (s *opensearchIndexer) handle(event *gen.PolicyEvaluationEvent) error {
	if !config.DcsServiceConfig().GetEnablePdpEventIndexing() {
		logger.Warnf("PDP event indexing is disabled, skipping eventId: %s", event.Header.Id)
		return nil
	}

	logger.Debugf("OpenSearch: Handling policy evaluation event: %s", event.Header.Id)

	jsonDoc, err := utils.ToPbJson(event, "")
	if err != nil {
		logger.Errorf("Failed to JSON serialize PDP event: %s", err.Error())
		return err
	}

	shardedName, ok := s.shardedNames.Load(config.DcsServiceConfig().GetPdpEventIndexName())
	if !ok {
		logger.Errorf("Index is not initialized into a sharded name")
		return errors.New("index not initialized")
	}

	indexReq := opensearchapi.IndexRequest{
		Index:      shardedName.(string),
		DocumentID: event.Header.Id,
		Body:       strings.NewReader(jsonDoc),
	}

	indexRes, err := indexReq.Do(context.Background(), s.client)
	if err != nil {
		logger.Errorf("Failed to index event: %s", err.Error())
		return err
	}

	logger.Debugf("Indexed document with status: %s", indexRes.Status())
	return nil
}

func (s *opensearchIndexer) buildOpenSearchClient() (*opensearch.Client, error) {
	cfg := config.DcsServiceConfig().GetOpensearchConfig()

	tlsConfig := tls.Config{}
	if cfg.GetAuthType() == gen.OpensearchIntegrationConfig_NONE {
		logger.Infof("Using Opensearch without authentication")
	} else {
		return nil, fmt.Errorf("unsupported auth type: %s", cfg.GetAuthType().String())
	}

	return opensearch.NewClient(opensearch.Config{
		Transport:  &http.Transport{TLSClientConfig: &tlsConfig},
		Addresses:  cfg.GetEndpoints(),
		MaxRetries: 3,
	})
}

func (s *opensearchIndexer) initOpenSearchIndex(name string) error {
	return utils.InvokeWithRetry(utils.RetryConfig{
		Count: 30,
		Sleep: time.Second * 1,
	}, func(n int) error {
		logger.Infof("Attempting to init opensearch index [retry=%d]", n)
		return s.initOpenSearchIndexInternal(name)
	})
}

func (s *opensearchIndexer) initOpenSearchIndexInternal(name string) error {
	if s.client == nil {
		return errors.New("client is nil")
	}

	shardableName := fmt.Sprintf("%s-%s", name, time.Now().Format("2006-01-02"))
	createIndexReq := opensearchapi.IndicesCreateRequest{
		Index: shardableName,
		Body: strings.NewReader(`{
			"settings": {},
			"mappings": {
				"properties": {
					"timestamp": {
						"type": "date",
						"format": "epoch_millis"
					}
				}
			}
		}`),
	}

	createIndexRes, err := createIndexReq.Do(context.Background(), s.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	s.shardedNames.Store(name, shardableName)
	logger.Debugf("Created OpenSearch index with name: %s status:%s",
		shardableName,
		createIndexRes.Status())
	return nil
}
