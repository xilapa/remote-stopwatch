package stopwatch

import (
	"testing"
	"time"

	assert "github.com/xilapa/go-tiny-projects/test-assertions"
)

type testObserver struct {
	times  []time.Duration
	resets []int
}

func (to *testObserver) HandleNewTime(t time.Duration) {
	to.times = append(to.times, t)
}

func (to *testObserver) HandleReset() {
	to.resets = append(to.resets, 0)
}

func newTestObserver() *testObserver {
	return &testObserver{
		times: make([]time.Duration, 0, 6),
	}
}

var _ Observer = (*testObserver)(nil)

func TestTimeCount(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	obs1 := newTestObserver()
	obs2 := newTestObserver()

	sw.Add(obs1)
	sw.Add(obs2)

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Stop()

	assert.Equal(t,
		obs1.times,
		obs2.times,
		"observers should have the same times")

	lastObservedTime := obs1.times[len(obs1.times)-1]

	assert.True(
		t,
		lastObservedTime >= 3*time.Second,
		"last observed time should be greater than 3 seconds",
	)

	assert.True(
		t,
		lastObservedTime == sw.stopTime,
		"last observed time should be equal to stop duration",
	)
}

func TestStopContinue(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	for i := 0; i < 5000; i++ {
		sw.Add(newTestObserver())
	}

	obs1 := newTestObserver()

	sw.Add(obs1)

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Stop()

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Stop()

	lastObservedTime := obs1.times[len(obs1.times)-1]

	assert.True(
		t,
		lastObservedTime >= 6*time.Second,
		"last observed time should be greater than 6 seconds",
	)

	assert.True(
		t,
		lastObservedTime == sw.stopTime,
		"last observed time should be equal to stop duration",
	)
}

func TestReset(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	for i := 0; i < 5000; i++ {
		sw.Add(newTestObserver())
	}

	obs1 := newTestObserver()

	sw.Add(obs1)

	sw.Start()
	<-time.After(3 * time.Second)

	sw.Reset()

	assert.True(
		t,
		sw.stopTime == 0,
		"last observed time should be equal to stop duration",
	)

	assert.True(
		t,
		len(obs1.resets) == 1,
		"observer should have been reset once",
	)
}

func TestResetStart(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	for i := 0; i < 5000; i++ {
		sw.Add(newTestObserver())
	}

	obs1 := newTestObserver()

	sw.Add(obs1)

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Reset()

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Stop()

	lastObservedTime := obs1.times[len(obs1.times)-1]

	assert.True(
		t,
		lastObservedTime >= 3*time.Second,
		"last observed time should be greater than 3 seconds",
	)

	assert.True(
		t,
		lastObservedTime == sw.stopTime,
		"last observed time should be equal to stop duration",
	)

	assert.True(
		t,
		len(obs1.resets) == 1,
		"observer should have been reset once",
	)
}

func TestStopTwice(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Stop()

	firstStopTime := sw.stopTime

	sw.Stop()

	assert.True(
		t,
		firstStopTime == sw.stopTime,
		"stop time should not change after second stop",
	)
}

func TestStartTwice(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()
	obs := newTestObserver()
	sw.Add(obs)

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Start()
	<-time.After(1 * time.Second)

	assert.True(t, sw.running, "stop watch should be running")

	lastObservedTime := obs.times[len(obs.times)-1]
	assert.True(t, lastObservedTime >= 4, "last observed time should be greater than 4 seconds")

	sw.Stop()
}

func TestResetTwice(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	obs := newTestObserver()
	sw.Add(obs)

	sw.Start()
	<-time.After(3 * time.Second)

	sw.Reset()
	sw.Reset()

	assert.True(t, sw.stopTime == 0, "stop time should be 0 after reset")
	assert.True(t, len(obs.resets) == 2, "observer should have 2 resets")
}

func TestRemoveObserver(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()
	assert.Equal(t, 0, sw.ObserversCount(), "should have 0 observers")

	obs1 := newTestObserver()
	obs2 := newTestObserver()

	sw.Add(obs1)
	sw.Add(obs2)
	assert.Equal(t, 2, sw.ObserversCount(), "should have 2 observers")

	sw.Remove(obs1)
	assert.Equal(t, 1, sw.ObserversCount(), "should have 1 observer")

	sw.Remove(obs2)
	assert.Equal(t, 0, sw.ObserversCount(), "should have 0 observers")
}
