package retry

import (
	"math"
	"math/rand"
	"time"
)

type BackoffStrategy interface {
	NextDelay(attempt int) time.Duration
}

type ExponentialBackoff struct {
	BaseDelay time.Duration //initial delay
	MaxDelay  time.Duration
	Jitter    bool
}

func (b ExponentialBackoff) NextDelay(attempt int) time.Duration {
	exp := float64(b.BaseDelay) * math.Pow(2, float64(attempt-1))
	delay := min(time.Duration(exp), b.MaxDelay)

	if b.Jitter {
		jitter := rand.Float64() * float64(delay) * 0.3
		delay += time.Duration(jitter)
	}

	return delay
}
