package wsobserver

import (
	"context"
	"fmt"
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
	fmt.Println("received time from stopwatch: ", t)
	select {
	// do not block if the channel is full
	case w.receivedTimes <- t:
		return
	default:
		return
	}
}

func (w *WsObserver) broadcast() {
	go func() {
		for t := range w.receivedTimes {
			// TODO: context
			err := w.c.Write(context.Background(), websocket.MessageText, []byte(t.String()))
			if err != nil {
				w.c.Close(websocket.StatusInternalError, "failed to write to websocket")
				return
			}
		}
	}()
}

func NewWebSocketObserver(ctx context.Context, c *websocket.Conn) stopwatch.Observer {
	wso := &WsObserver{
		receivedTimes: make(chan (time.Duration), 1),
		c:             c,
		ctx:           ctx,
	}
	wso.broadcast()
	return wso
}

var _ stopwatch.Observer = (*WsObserver)(nil)
