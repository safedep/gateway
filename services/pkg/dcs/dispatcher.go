package dcs

import (
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/safedep/gateway/services/pkg/common/logger"
	"github.com/safedep/gateway/services/pkg/common/messaging"
)

type eventSubscriptionHandler[T eventSubscriptionMessage] func(T) error
type eventSubscriptionDecoder[T eventSubscriptionMessage] func([]byte) (T, error)

type eventSubscription[T eventSubscriptionMessage] struct {
	name         string
	topic, group string
	decoder      eventSubscriptionDecoder[T]
	handler      eventSubscriptionHandler[T]
}

type eventSubscriptionMessage = proto.Message

var dispatcherWg sync.WaitGroup

// Register a subscriber to the messaging service and increment
// wait group. Perform generic event to subscriber specific type
// conversion and invoke subscriber business logic
func registerSubscriber[T eventSubscriptionMessage](msgService messaging.MessagingService,
	subscriber eventSubscription[T]) (messaging.MessagingQueueSubscription, error) {

	logger.Infof("Registering dispatcher name:%s topic:%s group:%s",
		subscriber.name, subscriber.topic, subscriber.group)

	sub, err := msgService.QueueSubscribe(subscriber.topic, subscriber.group, func(msg []byte) error {
		event, err := subscriber.decoder(msg)
		if err != nil {
			return err
		}

		return subscriber.handler(event)
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
