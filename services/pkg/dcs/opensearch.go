package dcs

import (
	"log"

	"github.com/safedep/gateway/services/gen"
	"github.com/safedep/gateway/services/pkg/common/config"
)

const (
	pdpEventIndexerName  = "PDP Events to OpenSearch Indexer"
	pdpEventIndexerGroup = "pdp-event-indexer-group"
)

type opensearchIndexer struct{}

func opensearchPdpEventIndexerSubscription() eventSubscription[gen.PolicyEvaluationEvent] {
	h := opensearchIndexer{}
	return h.pdpEventSubscription()
}

func (s *opensearchIndexer) pdpEventSubscription() eventSubscription[gen.PolicyEvaluationEvent] {
	cfg := config.PdpServiceConfig()

	return eventSubscription[gen.PolicyEvaluationEvent]{
		name:    pdpEventIndexerName,
		group:   pdpEventIndexerGroup,
		topic:   cfg.GetPublisherConfig().GetTopicNames().GetPolicyAudit(),
		handler: s.pdpEventHandler(),
	}
}

func (s *opensearchIndexer) pdpEventHandler() eventSubscriptionHandler[gen.PolicyEvaluationEvent] {
	return func(event *gen.PolicyEvaluationEvent) error {
		return s.handle(event)
	}
}

func (s *opensearchIndexer) handle(event *gen.PolicyEvaluationEvent) error {
	log.Printf("OpenSearch: Handling policy evaluation event")
	return nil
}
