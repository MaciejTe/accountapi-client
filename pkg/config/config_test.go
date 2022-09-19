package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	t.Parallel()
	testAddress := "https://address:8080"

	testCases := map[string]struct {
		Address        *string
		Timeout        time.Duration
		SkipVerify     bool
		ExpectedResult *Config
	}{
		"ok, default address": {
			Address:    nil,
			Timeout:    10000,
			SkipVerify: false,
			ExpectedResult: &Config{
				Address:    DefaultAddress,
				Timeout:    10000,
				SkipVerify: false,
			},
		},
		"ok, default timeout": {
			Address:    &testAddress,
			Timeout:    0,
			SkipVerify: false,
			ExpectedResult: &Config{
				Address:    testAddress,
				Timeout:    DefaultReqTimeout,
				SkipVerify: false,
			},
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(fmt.Sprintf("tc %s", name), func(t *testing.T) {
			result := NewConfig(tc.Address, tc.Timeout, tc.SkipVerify)
			assert.Equal(t, tc.ExpectedResult, result)
		})
	}
}
