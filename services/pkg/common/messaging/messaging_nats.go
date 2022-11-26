package messaging

import (
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"

	nats_proto "github.com/nats-io/nats.go/encoders/protobuf"
	config_api "github.com/safedep/gateway/services/gen"
	"github.com/safedep/gateway/services/pkg/common/logger"
)

type natsMessagingService struct {
	connection        *nats.Conn
	encodedConnection *nats.EncodedConn
}

// Coupled with protobuf encoder so expects protobuf serializable messages
func NewNatsMessagingService(cfg *config_api.MessagingAdapter) (MessagingService, error) {
	certs := nats.ClientCert(os.Getenv("SERVICE_TLS_CERT"), os.Getenv("SERVICE_TLS_KEY"))
	rootCA := nats.RootCAs(os.Getenv("SERVICE_TLS_ROOT_CA"))

	log.Printf("Initializing new nats connection with: %s", cfg)
	conn, err := nats.Connect(cfg.GetNats().Url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(5),
		nats.ReconnectWait(1*time.Second),
		certs, rootCA)

	if err != nil {
		return &natsMessagingService{}, err
	}

	err = conn.Flush()
	if err != nil {
		return &natsMessagingService{}, err
	}

	rtt, err := conn.RTT()
	if err != nil {
		return &natsMessagingService{}, err
	}

	log.Printf("NATS server connection initialized with RTT=%s", rtt)

	encodedConn, err := nats.NewEncodedConn(conn, nats_proto.PROTOBUF_ENCODER)
	if err != nil {
		return &natsMessagingService{}, err
	}

	return &natsMessagingService{connection: conn,
		encodedConnection: encodedConn}, nil
}

func (svc *natsMessagingService) QueueSubscribe(topic string, group string, handler MessageSubscriptionHandler) (MessagingQueueSubscription, error) {
	return svc.encodedConnection.QueueSubscribe(topic, group, func(m *nats.Msg) {
		err := handler(m.Data)
		if err != nil {
			logger.WithError(err).Errorf("message subscription handler failed")
		}
	})
}

func (svc *natsMessagingService) Publish(topic string, msg interface{}) error {
	return svc.encodedConnection.Publish(topic, msg)
}
