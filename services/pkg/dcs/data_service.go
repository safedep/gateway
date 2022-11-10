package dcs

import (
	"github.com/safedep/gateway/services/pkg/common/config"
	"github.com/safedep/gateway/services/pkg/common/db"
	"github.com/safedep/gateway/services/pkg/common/messaging"
)

type DataCollectionService struct {
	messagingService        messaging.MessagingService
	vulnerabilityRepository *db.VulnerabilityRepository
}

func NewDataCollectionService(msgService messaging.MessagingService,
	vRepo *db.VulnerabilityRepository) (*DataCollectionService, error) {

	return &DataCollectionService{messagingService: msgService,
		vulnerabilityRepository: vRepo}, nil
}

func (svc *DataCollectionService) Start() {
	registerSubscriber(svc.messagingService, sbomCollectorSubscription())
	registerSubscriber(svc.messagingService, vulnCollectorSubscription(svc.vulnerabilityRepository))

	if config.DcsServiceConfig().EnableOpensearchAdapter {
		svc.registerOpenSearchAdapter()
	}

	waitForSubscribers()
}

func (svc *DataCollectionService) registerOpenSearchAdapter() {
	registerSubscriber(svc.messagingService, opensearchPdpEventIndexerSubscription())
}
