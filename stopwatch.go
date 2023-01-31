package remotestopwatcher

import (
	"time"

	"github.com/jaevor/go-nanoid"
)

var nanoIdGen = mustCreateNanoIdGen()

const timeLoopDelay = time.Millisecond * 300

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
	Send(t time.Duration)
}

type StopWatch struct {
	id          string
	startTime   time.Time
	timeElapsed chan time.Duration
	stopChan    chan struct{}
	done        chan struct{}
	observers   []Observer
}

// TODO: use options pattern to configure a new stopwatcher

func NewStopWatcher() *StopWatch {
	return &StopWatch{
		id:          nanoIdGen(),
		startTime:   time.Time{},
		timeElapsed: make(chan time.Duration, 1),
		stopChan:    make(chan struct{}, 1),
		done:        make(chan struct{}),
		observers:   make([]Observer, 6),
	}
}

// timeLoop sends the elapsed time periodically
// to timeElapsed channel.
func (sw *StopWatch) timeLoop() {
	defer close(sw.done)

	select {
	// send the elapsed time periodically
	case <-time.After(timeLoopDelay):
		sw.timeElapsed <- time.Since(sw.startTime)

	// if stop is called, send the elapsed time to
	// the channel and stop the go function
	case <-sw.stopChan:
		sw.timeElapsed <- time.Since(sw.startTime)
		return
	}
}

// Start the stop watcher.
func (sw *StopWatch) Start() {
	sw.startTime = time.Now()

	// start time loop on another go routine
	go sw.timeLoop()

	// send the elapsed time to observers
	for t := range sw.timeElapsed {
		for i := range sw.observers {
			sw.observers[i].Send(t)
		}
		if sw.done == nil {
			return
		}
	}

	close(sw.timeElapsed)
}

// Stop the stop watcher, pausing the time.
func (sw *StopWatch) Stop() {
	close(sw.stopChan)
	<-sw.done
	sw.done = nil
}

// Continue the stop watcher from the current
// elapsed time.
func (sw *StopWatch) Continue() {
	sw.stopChan = make(chan struct{})
	sw.done = make(chan struct{})
	sw.timeElapsed = make(chan time.Duration)
	go sw.timeLoop()
}

// Reset the stopwatch time.
func (sw *StopWatch) Reset() {
	sw.Stop()
	sw.stopChan = make(chan struct{})
	sw.done = make(chan struct{})
	sw.timeElapsed = make(chan time.Duration)
	sw.Start()
}
