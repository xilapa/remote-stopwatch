package wsobserver

import (
	"context"
	"time"

	"github.com/xilapa/remote-stopwatch/stopwatch"
	"nhooyr.io/websocket"
)

type WsObserver struct {
	receivedTimes chan (time.Duration) // channel to receive time from the stopwatch
	c             *websocket.Conn
	ctx           context.Context
}

func (w *WsObserver) HandleNewTime(t time.Duration) {
	select {
	// do not block if the channel is full
	case w.receivedTimes <- t:
	default:
		return
	}
}

func (w *WsObserver) Broadcast() {
	for t := range w.receivedTimes {
		err := w.c.Write(w.ctx, websocket.MessageText, []byte(t.String()))
		if err != nil {
			w.c.Close(websocket.StatusInternalError, "failed to write to websocket")
			return
		}
	}
}

func NewWebSocketObserver(ctx context.Context, c *websocket.Conn) *WsObserver {
	return &WsObserver{
		receivedTimes: make(chan (time.Duration), 1),
		c:             c,
		ctx:           c.CloseRead(ctx),
	}
}

var _ stopwatch.Observer = (*WsObserver)(nil)
