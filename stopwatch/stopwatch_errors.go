package stopwatch

const (
	ErrStopWatchAlreadyRunningMsg = "stopwatch already running"
	ErrStopWatchNotRunningMsg     = "stopwatch not running"
)

type ErrStopWatchAlreadyRunning struct{}

func (e ErrStopWatchAlreadyRunning) Error() string {
	return ErrStopWatchAlreadyRunningMsg
}

type ErrStopWatchNotRunning struct{}

func (e ErrStopWatchNotRunning) Error() string {
	return ErrStopWatchNotRunningMsg
}

var _ error = (*ErrStopWatchAlreadyRunning)(nil)
var _ error = (*ErrStopWatchNotRunning)(nil)
