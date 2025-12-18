package rpc

import (
	"context"
	"time"
)

type Client interface {
	GetBlockHeight(ctx context.Context) (uint64, error)
	GetBalance(ctx context.Context, address string) (string, error)
	Close() error
	IsHealthy(ctx context.Context) bool
}

type Config struct {
	URL            string
	Timeout        time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	RequestTimeout time.Duration
}

func DefaultConfig(url string) Config {
	return Config{
		URL:            url,
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     time.Second,
		RequestTimeout: 10 * time.Second,
	}
}

func (c Config) WithTimeout(timeout time.Duration) Config {
	c.Timeout = timeout
	return c
}
func (c Config) WithMaxRetries(retries int) Config {
	c.MaxRetries = retries
	return c
}
