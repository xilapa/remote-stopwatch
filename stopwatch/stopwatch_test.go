package stopwatch

import (
	"testing"
	"time"

	assert "github.com/xilapa/go-tiny-projects/test-assertions"
)

type testObserver struct {
	times []time.Duration
}

func (to *testObserver) NewTime(t time.Duration) {
	to.times = append(to.times, t)
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

	sw.Continue()
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
	sw.Stop()

	sw.Reset()

	lastObservedTime := obs1.times[len(obs1.times)-1]

	assert.True(
		t,
		lastObservedTime == 0,
		"last observed time should be 0",
	)

	assert.True(
		t,
		lastObservedTime == sw.stopTime,
		"last observed time should be equal to stop duration",
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
	sw.Stop()

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
}

func TestStopTwice(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	sw.Start()
	<-time.After(3 * time.Second)
	err := sw.Stop()

	assert.NoError(t, err, "first stop should not return an error")

	firstStopTime := sw.stopTime

	err = sw.Stop()
	assert.Error(t, err, "second stop should return an error")
	assert.Equal(t, ErrStopWatchNotRunning{}, err)

	assert.True(
		t,
		firstStopTime == sw.stopTime,
		"stop time should not change after second stop",
	)
}

func TestStartTwice(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	err := sw.Start()
	assert.NoError(t, err, "first start should not return an error")

	err = sw.Start()
	assert.Error(t, err, "second start should return an error")
	assert.Equal(t, ErrStopWatchAlreadyRunning{}, err)

	sw.Stop()
}

func TestContinueTwice(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Stop()

	err := sw.Continue()
	assert.NoError(t, err, "first continue should not return an error")

	err = sw.Continue()
	assert.Error(t, err, "second continue should return an error")
	assert.Equal(t, ErrStopWatchAlreadyRunning{}, err)

	sw.Stop()
}

func TestResetTwice(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	sw.Start()
	<-time.After(3 * time.Second)
	sw.Stop()

	sw.Reset()
	sw.Reset()

	assert.True(t, sw.stopTime == 0, "stop time should be 0 after reset")
}

func TestContinueWithoutStop(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	sw.Start()
	<-time.After(3 * time.Second)

	err := sw.Continue()
	assert.Error(t, err, "continue should return an error")
	assert.Equal(t, ErrStopWatchAlreadyRunning{}, err)

	sw.Stop()
}

func TestContinueWithoutStart(t *testing.T) {
	t.Parallel()

	sw := NewStopWatch()

	err := sw.Continue()
	assert.NoError(t, err, "continue should return an error")
	assert.Equal(t, sw.running, true)

	sw.Stop()
}
