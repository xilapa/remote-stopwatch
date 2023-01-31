package remotestopwatch

import (
	"testing"
	"time"

	assert "github.com/xilapa/go-tiny-projects/test-assertions"
)

type testObserver struct {
	times []time.Duration
}

func (to *testObserver) Send(t time.Duration) {
	to.times = append(to.times, t)
}

func newTestObserver() *testObserver {
	return &testObserver{
		times: make([]time.Duration, 0, 6),
	}
}

var _ Observer = (*testObserver)(nil)

// TODO: table driven tests
// timeCountData := map[string]struct{
// 	total int
// }

func TestTimeCount(t *testing.T) {
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

	// TODO: make an assert true on tiny-projects
	assert.Equal(t,
		lastObservedTime >= 3,
		true,
		"last observed time should be greater than 3 seconds")
}

// TODO: test for stop-> continue
// TODO: test for stop-> reset -> start
// TODO: test for stop twice
// TODO: test for start twice
// TODO: test for continue twice
// TODO: test for reset twice
