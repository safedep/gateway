package dcs

import (
	"sync"

	"github.com/safedep/gateway/services/pkg/common/logger"
	"github.com/safedep/gateway/services/pkg/common/messaging"
	"github.com/safedep/gateway/services/pkg/common/utils"
)

type eventSubscriptionHandler[T any] func(*T) error

type eventSubscription[T any] struct {
	name         string
	topic, group string
	handler      eventSubscriptionHandler[T]
}

var dispatcherWg sync.WaitGroup

// Register a subscriber to the messaging service and increment
// wait group. Perform generic event to subscriber specific type
// conversion and invoke subscriber business logic
func registerSubscriber[T any](msgService messaging.MessagingService,
	subscriber eventSubscription[T]) (messaging.MessagingQueueSubscription, error) {

	logger.Infof("Registering dispatcher name:%s topic:%s group:%s",
		subscriber.name, subscriber.topic, subscriber.group)

	sub, err := msgService.QueueSubscribe(subscriber.topic, subscriber.group, func(msg interface{}) {
		var event T
		if err := utils.MapStruct(msg, &event); err == nil {
			subscriber.handler(&event)
		} else {
			logger.Infof("Error creating a domain event of type T from event msg: %v", err)
		}
	})

	if err != nil {
		logger.Errorf("Error registering queue subscriber: %v", err)
	} else {
		dispatcherWg.Add(1)
	}

	return sub, err
}

func waitForSubscribers() {
	logger.Infof("Dispatcher waiting for queue subscriptions to close")
	dispatcherWg.Wait()
}
