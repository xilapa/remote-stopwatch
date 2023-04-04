package stopwatch

import (
	"sync"
	"time"

	"github.com/jaevor/go-nanoid"
)

var nanoIdGen = mustCreateNanoIdGen()

const timeLoopDelay = time.Millisecond * 150

func mustCreateNanoIdGen() func() string {
	nanoIdGen, err := nanoid.Standard(21)
	if err != nil {
		panic(err)
	}
	return nanoIdGen
}

// Observer is something that wants to
// listens to the stopwatch.
type Observer interface {
	HandleNewTime(t time.Duration)
}

type StopWatch struct {
	Id           string             // id of the stop watch
	CurrentTime  time.Duration      // current time of the stop watch
	startTime    time.Time          // start time of the stop watch
	timeElapsed  chan time.Duration // channel to send the elapsed time
	stopChan     chan struct{}      // channel that indicates the intention to stop the StopWatch
	timeLoopDone chan struct{}      // channel that indicates the timeLoop is done/stopped
	running      bool               // indicates the StopWatch is running
	observers    []Observer         // observers that want to listen to the StopWatch
	stopTime     time.Duration      // the duration of the StopWatch when it is stopped
	mtx          sync.Mutex         // mutex to protect the observers
}

// TODO: use options pattern to configure a new stopwatcher

func NewStopWatch() *StopWatch {
	return &StopWatch{
		Id:           nanoIdGen(),
		startTime:    time.Time{},
		timeElapsed:  make(chan time.Duration, 1),
		stopChan:     make(chan struct{}, 1),
		timeLoopDone: make(chan struct{}, 1),
		observers:    make([]Observer, 0, 6),
		stopTime:     time.Duration(0),
	}
}

// timeLoop sends the elapsed time periodically
// to timeElapsed channel.
func (sw *StopWatch) timeLoop() {
	defer close(sw.timeLoopDone)

	for {
		select {
		// send the elapsed time periodically
		case <-time.After(timeLoopDelay):
			sw.CurrentTime = time.Since(sw.startTime)
			sw.timeElapsed <- sw.CurrentTime

		// if stop is called, save the elapsed time and
		// send it to the channel
		// then stop the go function
		case <-sw.stopChan:
			sw.stopTime = time.Since(sw.startTime)
			sw.CurrentTime = sw.stopTime
			sw.timeElapsed <- sw.stopTime
			return
		}
	}
}

func (sw *StopWatch) sendTimeToObservers(t time.Duration) {
	for o := range sw.observers {
		sw.observers[o].HandleNewTime(t)
	}
}

func (sw *StopWatch) changeRunningState(running bool) {
	sw.mtx.Lock()
	sw.running = running
	sw.mtx.Unlock()
}

// startFrom starts the stopwatch from a given time.
func (sw *StopWatch) startFrom(t time.Time) {
	sw.changeRunningState(true)

	// the start time
	sw.startTime = t

	// start time loop on another go routine
	go sw.timeLoop()

	// listen to timeElapsed channel on another go routine
	go func() {
		defer close(sw.timeElapsed)
		// send the elapsed time to observers
		for t := range sw.timeElapsed {
			sw.sendTimeToObservers(t)

			// if stop is called, send the last timeElapsed
			// and then stop the go function
			select {
			case <-sw.timeLoopDone:
				sw.sendTimeToObservers(sw.stopTime)
				return
			default:
				continue
			}
		}
	}()
}

// resetChannels resets the channels.
func (sw *StopWatch) resetChannels() {
	sw.stopChan = make(chan struct{})
	sw.timeElapsed = make(chan time.Duration)
	sw.timeLoopDone = make(chan struct{})
}

// Start the stop watcher.
func (sw *StopWatch) Start() error {
	if sw.running {
		return ErrStopWatchAlreadyRunning{}
	}
	sw.startFrom(time.Now())
	return nil
}

// Stop the stop watcher, pausing the time.
func (sw *StopWatch) Stop() error {
	if !sw.running {
		return ErrStopWatchNotRunning{}
	}
	close(sw.stopChan)
	<-sw.timeElapsed
	sw.resetChannels()
	sw.changeRunningState(false)
	return nil
}

// Continue the stop watcher from the current
// elapsed time.
func (sw *StopWatch) Continue() error {
	if sw.running {
		return ErrStopWatchAlreadyRunning{}
	}

	sw.startFrom(time.Now().Add(-sw.stopTime))
	return nil
}

// Reset the stopwatch time.
func (sw *StopWatch) Reset() {
	// if the stopwatch is running, stop it
	if sw.running {
		sw.Stop()
	}
	sw.stopTime = time.Duration(0)
	sw.sendTimeToObservers(sw.stopTime)
}

// Add a new observer to the stopwatch.
func (sw *StopWatch) Add(o Observer) {
	sw.mtx.Lock()
	defer sw.mtx.Unlock()

	sw.observers = append(sw.observers, o)
}

// Remove an observer from the stopwatch.
func (sw *StopWatch) Remove(o Observer) {
	sw.mtx.Lock()
	defer sw.mtx.Unlock()

	for i, obs := range sw.observers {
		if obs == o {
			sw.observers = append(sw.observers[:i], sw.observers[i+1:]...)
			return
		}
	}
}

func (sw *StopWatch) ObserversCount() int {
	return len(sw.observers)
}
