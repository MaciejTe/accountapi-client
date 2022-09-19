package config

import (
	"time"
)

var (
	DefaultAddress    = "http://localhost:8080"
	DefaultReqTimeout = time.Duration(5) * time.Second
)

// Config conveys account API client configuration
type Config struct {
	// Account API host
	Address string
	// Request timeout
	Timeout time.Duration
	// do not verify server's certificate; warning: do not skip verification in production
	SkipVerify bool
}

// NewConfig returns pointer to the account API configuration structure
func NewConfig(address *string, timeout time.Duration, skipVerify bool) *Config {
	// TODO: Config struct validation would be nice, but I'm not sure what "client validation" means for reviewers
	if address == nil || *address == "" {
		address = &DefaultAddress
	}

	if timeout == 0 {
		timeout = DefaultReqTimeout
	}

	return &Config{
		Address:    *address,
		Timeout:    timeout,
		SkipVerify: skipVerify,
	}
}
