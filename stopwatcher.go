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

type Observer interface {
	Send(t time.Duration)
}

type StopWatcher struct {
	id          string
	startTime   time.Time
	timeElapsed chan time.Duration
	stopChan    chan struct{}
	done        chan struct{}
	observers   []Observer
}

func NewStopWatcher() *StopWatcher {
	return &StopWatcher{
		id:          nanoIdGen(),
		startTime:   time.Time{},
		timeElapsed: make(chan time.Duration, 1),
		stopChan:    make(chan struct{}, 1),
		done:        make(chan struct{}),
		observers:   make([]Observer, 6),
	}
}

func (sw *StopWatcher) timeLoop() {
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

func (sw *StopWatcher) Start() {
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

func (sw *StopWatcher) Stop() {
	close(sw.stopChan)
	<-sw.done
	sw.done = nil
}

func (sw *StopWatcher) Continue() {
	sw.stopChan = make(chan struct{})
	sw.done = make(chan struct{})
	sw.timeElapsed = make(chan time.Duration)
	go sw.timeLoop()
}

func (sw *StopWatcher) Reset() {
	sw.Stop()
	sw.stopChan = make(chan struct{})
	sw.done = make(chan struct{})
	sw.timeElapsed = make(chan time.Duration)
	sw.Start()
}
