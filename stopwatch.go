package remotestopwatch

import (
	"sync"
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
	NewTime(t time.Duration)
}

type StopWatch struct {
	id          string
	startTime   time.Time
	timeElapsed chan time.Duration
	stopChan    chan struct{}
	done        chan struct{}
	observers   []Observer
	mtx         sync.Mutex
}

// TODO: use options pattern to configure a new stopwatcher

func NewStopWatch() *StopWatch {
	return &StopWatch{
		id:          nanoIdGen(),
		startTime:   time.Time{},
		timeElapsed: make(chan time.Duration, 1),
		stopChan:    make(chan struct{}, 1),
		done:        make(chan struct{}, 1),
		observers:   make([]Observer, 0, 6),
	}
}

// timeLoop sends the elapsed time periodically
// to timeElapsed channel.
func (sw *StopWatch) timeLoop() {
	defer close(sw.done)

	for {
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
}

// Start the stop watcher.
func (sw *StopWatch) Start() {

	// TODO: create an internal start function that receives
	// the start time

	sw.startTime = time.Now()

	// start time loop on another go routine
	go sw.timeLoop()

	// listen to timeElapsed channel on another go routine
	go func() {
		defer close(sw.timeElapsed)
		// send the elapsed time to observers
		for t := range sw.timeElapsed {
			for i := range sw.observers {
				sw.observers[i].NewTime(t)
			}

			// if stop is called, stop the go function
			select {
			case <-sw.stopChan:
				return
			default:
				continue
			}
		}
	}()
}

// Stop the stop watcher, pausing the time.
func (sw *StopWatch) Stop() {
	// TODO: what happens if stop is called twice?
	close(sw.stopChan)
	<-sw.done
}

// Continue the stop watcher from the current
// elapsed time.
func (sw *StopWatch) Continue() {
	sw.stopChan = make(chan struct{})
	sw.done = make(chan struct{})
	sw.timeElapsed = make(chan time.Duration)
	// TODO: continue should call start
	go sw.timeLoop()
}

// Reset the stopwatch time.
func (sw *StopWatch) Reset() {
	sw.Stop()
	// TODO: reset should call continue
	sw.stopChan = make(chan struct{})
	sw.done = make(chan struct{})
	sw.timeElapsed = make(chan time.Duration)
	sw.Start()
}

// Add a new observer to the stopwatch.
func (sw *StopWatch) Add(o Observer) {
	sw.mtx.Lock()
	defer sw.mtx.Unlock()

	sw.observers = append(sw.observers, o)
}
