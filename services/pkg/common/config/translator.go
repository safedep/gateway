package config

import "github.com/safedep/gateway/services/gen"

type PullTranslator[T any] interface {
	Translate(gen.GatewayConfiguration) (T, error)
}

type PushTranslator[T any] interface {
	RegisterReceiver(func(T, error) error)
}
