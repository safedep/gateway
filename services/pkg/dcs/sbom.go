package dcs

import (
	"log"

	"github.com/safedep/gateway/services/gen"
	"github.com/safedep/gateway/services/pkg/common/config"
	"google.golang.org/protobuf/proto"
)

const (
	sbomCollectorGroupName = "sbom-collector-group"
	sbomCollectorName      = "SBOM Data Collector"
)

type sbomCollector struct{}

func sbomCollectorSubscription() eventSubscription[*gen.TapArtefactRequestEvent] {
	h := &sbomCollector{}
	return h.subscription()
}

func (s *sbomCollector) subscription() eventSubscription[*gen.TapArtefactRequestEvent] {
	cfg := config.TapServiceConfig()

	return eventSubscription[*gen.TapArtefactRequestEvent]{
		name:  sbomCollectorName,
		group: sbomCollectorGroupName,
		topic: cfg.GetPublisherConfig().GetTopicNames().GetUpstreamRequest(),
		decoder: func(b []byte) (*gen.TapArtefactRequestEvent, error) {
			var event gen.TapArtefactRequestEvent
			err := proto.Unmarshal(b, &event)
			return &event, err
		},
		handler: s.handler(),
	}
}

func (s *sbomCollector) handler() eventSubscriptionHandler[*gen.TapArtefactRequestEvent] {
	return func(event *gen.TapArtefactRequestEvent) error {
		return s.handle(event)
	}
}

func (s *sbomCollector) handle(event *gen.TapArtefactRequestEvent) error {
	log.Printf("SBOM collector - Handling artefact: %v", event.Data)
	return nil
}
