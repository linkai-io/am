package retrier

import (
	"time"

	retry "github.com/avast/retry-go"
)

// Retry default retrier, retries 10 times, with exponential back off
func Retry(retryFn retry.RetryableFunc) error {
	return retry.Do(retryFn)
}

func RetryUnless(retryFn retry.RetryableFunc, errType error) error {
	return retry.Do(
		retryFn,
		retry.RetryIf(func(err error) bool {
			if err == errType {
				return false
			}
			return true
		}))
}

// RetryAttempts retries attempts times
func RetryAttempts(retryFn retry.RetryableFunc, attempts uint) error {
	return retry.Do(retryFn, retry.Attempts(attempts))
}

// RetryUntil simple retrier until time runs out, checks every tick, does not do exponential back off
func RetryUntil(retryFn retry.RetryableFunc, until time.Duration, tick time.Duration) error {
	ticker := time.NewTicker(tick)
	timer := time.NewTimer(until)
	defer ticker.Stop()
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			if err := retryFn(); err == nil {
				return nil
			}
		case <-timer.C:
			return retryFn()
		}
	}
}
