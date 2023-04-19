package stopwatch

import (
	"sync"
	"time"

	"github.com/jaevor/go-nanoid"
)

var nanoIdGen = mustCreateNanoIdGen()

func mustCreateNanoIdGen() func() string {
	nanoIdGen, err := nanoid.Standard(21)
	if err != nil {
		panic(err)
	}
	return nanoIdGen
}

type StopWatch struct {
	Id               string             // id of the stopwatch
	CurrentTime      time.Duration      // current time of the stopwatch
	IdleSince        time.Time          // time when the stopwatch start to be idle
	startTime        time.Time          // start time of the stopwatch
	timeElapsed      chan time.Duration // channel to send the elapsed time
	stopChan         chan struct{}      // channel that indicates the intention to stop the StopWatch
	timeLoopDone     chan struct{}      // channel that indicates the timeLoop is done/stopped
	running          bool               // indicates the StopWatch is running
	observers        []Observer         // observers that want to listen to the StopWatch
	stopTime         time.Duration      // the duration of the StopWatch when it is stopped
	timeLoopInterval time.Duration      // the interval of the timeLoop
	mtx              sync.Mutex         // mutex to protect the observers
}

func NewStopWatch(opts ...StopwatchOptions) *StopWatch {
	sw := &StopWatch{
		Id:               nanoIdGen(),
		startTime:        time.Time{},
		timeElapsed:      make(chan time.Duration, 1),
		stopChan:         make(chan struct{}, 1),
		timeLoopDone:     make(chan struct{}, 1),
		observers:        make([]Observer, 0, 6),
		stopTime:         time.Duration(0),
		timeLoopInterval: time.Millisecond * 150,
	}

	for i := range opts {
		opts[i](sw)
	}

	return sw
}

// timeLoop sends the elapsed time periodically
// to timeElapsed channel.
func (sw *StopWatch) timeLoop() {
	defer close(sw.timeLoopDone)

	for {
		select {
		// send the elapsed time periodically
		case <-time.After(sw.timeLoopInterval):
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
	for idx := range sw.observers {
		sw.sendTimeToObserserver(idx, t)
	}
}

// sendTimeToObserserver sends the elapsed time to an observer.
// If the observer is removed during sending the time, it will
// recover from panic.
func (sw *StopWatch) sendTimeToObserserver(idx int, t time.Duration) {
	defer recover()
	sw.observers[idx].HandleNewTime(t)
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

// Start the stopwatch or continue if the stopwatch is stopped.
func (sw *StopWatch) Start() {
	if sw.running {
		return
	}
	sw.startFrom(time.Now().Add(-sw.stopTime))
}

// Stop the stopwatch, pausing the time.
func (sw *StopWatch) Stop() {
	if !sw.running {
		return
	}
	close(sw.stopChan)
	<-sw.timeElapsed
	sw.resetChannels()
	sw.changeRunningState(false)
}

// Reset the stopwatch time.
func (sw *StopWatch) Reset() {
	// if the stopwatch is running, stop it
	if sw.running {
		sw.Stop()
	}
	sw.stopTime = time.Duration(0)
	for idx := range sw.observers {
		sw.sendResetToObserver(idx)
	}
}

// sendResetToObserver sends the reset signal to an observer.
// If the observer is removed during sending the reset signal,
// it will recover from panic.
func (sw *StopWatch) sendResetToObserver(idx int) {
	defer recover()
	sw.observers[idx].HandleReset()
}

// Add a new observer to the stopwatch.
func (sw *StopWatch) Add(o Observer) {
	sw.mtx.Lock()
	defer sw.mtx.Unlock()

	sw.observers = append(sw.observers, o)
	sw.IdleSince = time.Time{}
}

// Remove an observer from the stopwatch.
func (sw *StopWatch) Remove(o Observer) {
	sw.mtx.Lock()
	defer sw.mtx.Unlock()

	for i := range sw.observers {
		if sw.observers[i] == o {
			sw.observers = append(sw.observers[:i], sw.observers[i+1:]...)
		}
	}

	if len(sw.observers) == 0 {
		sw.IdleSince = time.Now()
	}
}

// ObserversCount returns the current number of observers.
func (sw *StopWatch) ObserversCount() int {
	return len(sw.observers)
}
