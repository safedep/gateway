package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvokeWithRetry(t *testing.T) {
	cases := []struct {
		name string
		c    RetryConfig
		f    RetriableFunc
		err  error
	}{
		{
			"Must fail with zero Count",
			RetryConfig{
				Count: 0,
				Sleep: 0,
			},
			func(n int) error { return nil },
			errInvalidRetryCount,
		},
		{
			"Must fail with zero Sleep",
			RetryConfig{
				Count: 1,
				Sleep: 0,
			},
			func(n int) error { return nil },
			errInvalidSleepDuration,
		},
		{
			"Retry is successful immediately",
			RetryConfig{
				Count: 1,
				Sleep: time.Millisecond * 2,
			},
			func(n int) error {
				return nil
			},
			nil,
		},
		{
			"Retry is successful after 5 attempts",
			RetryConfig{
				Count: 5,
				Sleep: time.Millisecond * 1,
			},
			func(n int) error {
				if n < 5 {
					return errors.New("< 5")
				}

				return nil
			},
			nil,
		},
		{
			"Retry is never successful",
			RetryConfig{
				Count: 10,
				Sleep: time.Millisecond * 1,
			},
			func(n int) error {
				return errors.New("err")
			},
			errors.New("err"),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := InvokeWithRetry(test.c, test.f)
			if test.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, test.err.Error(), err.Error())
			}
		})
	}
}
