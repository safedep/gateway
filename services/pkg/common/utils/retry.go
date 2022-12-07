package utils

import (
	"errors"
	"time"
)

var (
	errInvalidRetryCount    = errors.New("invalid retry count")
	errInvalidSleepDuration = errors.New("must have a valid sleep")
)

type RetryFuncArg struct {
	// Total retries to be executed
	Total int

	// Current try count starting with 1
	Current int
}

type RetriableFunc func(arg RetryFuncArg) error

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
		err = f(RetryFuncArg{Total: config.Count, Current: (i + 1)})
		if err == nil {
			break
		}

		time.Sleep(config.Sleep)
	}

	return err
}
