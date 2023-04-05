package stopwatchclient

import (
	"context"
	"fmt"
	"time"

	"github.com/xilapa/remote-stopwatch/stopwatch"
	"nhooyr.io/websocket"
)

type StopWatchWSClient struct {
	receivedTimes chan (time.Duration) // channel to receive time from the stopwatch
	c             *websocket.Conn      // websocket connection
	ctx           context.Context      // context relate to http connection
	sw            *stopwatch.StopWatch // stopwatch reference
}

func NewWebSocketClient(ctx context.Context, c *websocket.Conn) *StopWatchWSClient {
	return &StopWatchWSClient{
		receivedTimes: make(chan (time.Duration), 1),
		c:             c,
		ctx:           c.CloseRead(ctx),
	}
}

// HandleNewTime is called by the stopwatch when a new time is received.
// The method sends the time to a channel without blocking.
func (w *StopWatchWSClient) HandleNewTime(t time.Duration) {
	select {
	// do not block if the channel is full
	case w.receivedTimes <- t:
	default:
		return
	}
}

// broadcast sends the received time to the websocket client.
func (w *StopWatchWSClient) broadcast() {
	for t := range w.receivedTimes {
		err := w.c.Write(w.ctx, websocket.MessageText, []byte(fmt.Sprintf("%d", t.Milliseconds())))
		if err != nil {
			w.stopClient()
			return
		}
	}
}

// stopClient closes the websocket connection and removes the client from the stopwatch.
func (w *StopWatchWSClient) stopClient() {
	w.c.Close(websocket.StatusInternalError, "failed to write to websocket")
	w.sw.Remove(w)
	close(w.receivedTimes)
}

// Handle the stopwatch through the websocket connection.
// The method only returns when the websocket connection is closed.
func (w *StopWatchWSClient) Handle(sw *stopwatch.StopWatch) {
	w.sw = sw
	w.sw.Add(w)
	w.broadcast()
}

var _ stopwatch.Observer = (*StopWatchWSClient)(nil)
