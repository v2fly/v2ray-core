package retry

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"math"
	"math/rand"
	"time"
)

var ErrRetryFailed = newError("all retry attempts failed")

// Strategy is a way to retry on a specific function.
type Strategy interface {
	// On performs a retry on a specific function, until it doesn't return any error.
	On(func() error) error
}

type retryer struct {
	totalAttempt int
	nextDelay    func() uint32
}

// On implements Strategy.On.
func (r *retryer) On(method func() error) error {
	attempt := 0
	accumulatedError := make([]error, 0, r.totalAttempt)
	for attempt < r.totalAttempt {
		err := method()
		if err == nil {
			return nil
		}
		numErrors := len(accumulatedError)
		if numErrors == 0 || err.Error() != accumulatedError[numErrors-1].Error() {
			accumulatedError = append(accumulatedError, err)
		}
		delay := r.nextDelay()
		time.Sleep(time.Duration(delay) * time.Millisecond)
		attempt++
	}
	return newError(accumulatedError).Base(ErrRetryFailed)
}

// Timed returns a retry strategy with fixed interval.
func Timed(attempts int, delay uint32) Strategy {
	return &retryer{
		totalAttempt: attempts,
		nextDelay: func() uint32 {
			return delay
		},
	}
}

func ExponentialBackoff(attempts int, delay uint32) Strategy {
	nextDelay := uint32(0)
	return &retryer{
		totalAttempt: attempts,
		nextDelay: func() uint32 {
			r := nextDelay
			nextDelay += delay
			return r
		},
	}
}

// ExponentialBackoffWithJitter return a retry exponential backoff strategy with jitter
// http://www.awsarchitectureblog.com/2015/03/backoff.html
func ExponentialBackoffWithJitter(attempts int, delay uint32) Strategy {
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	capLevel := delay * 5
	epoch := 0
	return &retryer{
		totalAttempt: attempts,
		nextDelay: func() uint32 {
			temp := math.Min(float64(capLevel), float64(delay)*math.Exp2(float64(epoch)))
			ri := int64(temp / 2)
			epoch += 1
			jitter := rr.Int63n(ri)
			return uint32(ri + jitter)
		},
	}
}
