package backoff

import (
	"math"
	"math/rand"
	"time"
)

// FullJitter computes an exponential backoff with full jitter.
// attemptIndex starts at 0. maxDelay is the maximum sleep.
func FullJitter(attemptIndex int, baseDelay time.Duration, multiplier float64, maxDelay time.Duration) time.Duration {
	exponential := time.Duration(float64(baseDelay) * math.Pow(multiplier, float64(attemptIndex)))
	if exponential > maxDelay {
		exponential = maxDelay
	}
	randomNanos := rand.Int63n(int64(exponential) + 1) // [0, exponential]
	return time.Duration(randomNanos)
}
