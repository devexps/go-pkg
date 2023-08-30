package ratelimiter

import (
	"errors"
)

var (
	// ErrLimitExceed is returned when the rate limiter is
	// triggered and the request is rejected due to limit exceeded.
	ErrLimitExceed = errors.New("rate limit exceeded")
)

// DoneFunc is done function.
type DoneFunc func(DoneInfo)

// DoneInfo is done info.
type DoneInfo struct {
	Err error
}

// RateLimiter is a rate limiter.
type RateLimiter interface {
	Allow() (DoneFunc, error)
}
