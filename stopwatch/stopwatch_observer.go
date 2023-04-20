package stopwatch

import "time"

// Observer is something that wants to
// listens to the stopwatch.
type Observer interface {
	// HandleNewTime is called when the stopwatch report a new time frame.
	// This method should not block due to the fact that the stopwatch
	// is blocked until all observers are done. Also, this method is called
	// with high frequency, so it should be as fast as possible.
	HandleNewTime(t time.Duration)

	// HandleReset is called when the stopwatch is reset.
	// This method should block until the observer is done. Because the
	// observers must know that the stopwatch is reset.
	HandleReset()

	// HandleObserverCountChange is called when the number of observers
	// changes. This method should block.
	HandleObserverCountChange(count int)
}
