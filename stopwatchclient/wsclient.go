package stopwatchclient

import (
	"context"
	"time"

	"github.com/xilapa/remote-stopwatch/stopwatch"
	"nhooyr.io/websocket"
)

type StopWatchWSClient struct {
	receivedTimes chan (time.Duration) // channel to receive time from the stopwatch
	c             *websocket.Conn
	ctx           context.Context
}

func (w *StopWatchWSClient) HandleNewTime(t time.Duration) {
	select {
	// do not block if the channel is full
	case w.receivedTimes <- t:
	default:
		return
	}
}

func (w *StopWatchWSClient) Broadcast() {
	for t := range w.receivedTimes {
		err := w.c.Write(w.ctx, websocket.MessageText, []byte(t.String()))
		if err != nil {
			w.c.Close(websocket.StatusInternalError, "failed to write to websocket")
			return
		}
	}
}

func NewWebSocketClient(ctx context.Context, c *websocket.Conn) *StopWatchWSClient {
	return &StopWatchWSClient{
		receivedTimes: make(chan (time.Duration), 1),
		c:             c,
		ctx:           c.CloseRead(ctx),
	}
}

var _ stopwatch.Observer = (*StopWatchWSClient)(nil)
