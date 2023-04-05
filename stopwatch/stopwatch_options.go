package stopwatch

import "time"

type StopwatchOptions func(*StopWatch)

func WithTimeLoopInterval(t time.Duration) StopwatchOptions {
	return func(sw *StopWatch) {
		sw.timeLoopInterval = t
	}
}
