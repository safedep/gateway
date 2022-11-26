package utils

import (
	"errors"
	"time"
)

var (
	errInvalidRetryCount    = errors.New("invalid retry count")
	errInvalidSleepDuration = errors.New("must have a valid sleep")
)

type RetriableFunc func(retryN int) error

type RetryConfig struct {
	Count int
	Sleep time.Duration
}

func InvokeWithRetry(config RetryConfig, f RetriableFunc) error {
	if config.Count <= 0 {
		return errInvalidRetryCount
	}

	now := time.Now()
	if now.Add(config.Sleep) == now {
		return errInvalidSleepDuration
	}

	var err error
	for i := 0; i < config.Count; i += 1 {
		err = f(i + 1)
		if err == nil {
			break
		}

		time.Sleep(config.Sleep)
	}

	return err
}
