package stopwatchclient

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xilapa/remote-stopwatch/stopwatch"
	"nhooyr.io/websocket"
)

var (
	startBytes = []byte("start")
	stopBytes  = []byte("stop")
	resetBytes = []byte("reset")
)

type StopWatchWSClient struct {
	msgs        chan (string)        // channel to receive time from the stopwatch
	c           *websocket.Conn      // websocket connection
	ctx         context.Context      // context related to http connection
	sw          *stopwatch.StopWatch // stopwatch reference
	err         chan error           // channel to report erros on the websocket connection
	readStopped chan struct{}        // signal that read from websocket connection is stopped
}

func NewWebSocketClient(ctx context.Context, c *websocket.Conn) *StopWatchWSClient {
	return &StopWatchWSClient{
		msgs:        make(chan (string), 1),
		err:         make(chan error, 1),
		readStopped: make(chan struct{}),
		c:           c,
		ctx:         ctx,
	}
}

// HandleNewTime is called by the stopwatch when a new time is received.
// The method sends the time to a channel without blocking.
func (w *StopWatchWSClient) HandleNewTime(t time.Duration) {
	msg := fmt.Sprintf("time:%d", t.Milliseconds())
	select {
	// do not block if the channel is full
	case w.msgs <- msg:
	default:
		return
	}
}

// HandleReset is called by the stopwatch when the stopwatch is reset.
// The method sends a zero time to the messages channel, blocking.
func (w *StopWatchWSClient) HandleReset() {
	w.msgs <- "time:0"
}

// Handle the stopwatch through the websocket connection.
// The method only returns when the websocket connection is closed.
func (w *StopWatchWSClient) Handle(sw *stopwatch.StopWatch) {
	w.sw = sw
	w.sw.Add(w)

	// start reading and send commands to the websocket connection
	go w.read()
	w.send()

	// when send() returns, stop this websocket client
	w.stopClient()
}

func (w *StopWatchWSClient) read() {
	defer close(w.readStopped)
	for {
		select {
		// case an error is reported by send(), stop read() and return the error
		case err := <-w.err:
			w.err <- err
			return
		default:
			t, b, err := w.c.Read(w.ctx)
			// report error and return
			if err != nil {
				w.err <- err
				return
			}

			if t != websocket.MessageText {
				continue
			}

			w.executeStopwatchCmmd(b)
		}
	}
}

// executeStopwatchCmmd executes the command received from the websocket client.
func (w *StopWatchWSClient) executeStopwatchCmmd(b []byte) {
	switch {
	case bytes.Equal(b, startBytes):
		w.sw.Start()
	case bytes.Equal(b, stopBytes):
		w.sw.Stop()
	case bytes.Equal(b, resetBytes):
		w.sw.Reset()
	}
}

// send messages to the websocket client.
func (w *StopWatchWSClient) send() {
	for msg := range w.msgs {
		select {
		// case an error is reported by read(), stop send() and return the error
		case err := <-w.err:
			w.err <- err
			return
		default:
			err := w.c.Write(w.ctx, websocket.MessageText, []byte(msg))
			// report error and return
			if err != nil {
				w.err <- err
				return
			}
		}
	}
}

// stopClient closes the websocket connection and removes the client from the stopwatch.
func (w *StopWatchWSClient) stopClient() {
	// read the error
	err := <-w.err
	// wait for read() to return
	<-w.readStopped
	// close the websocket connection
	w.c.Close(websocket.StatusInternalError, err.Error())
	// remove the client from the stopwatch
	w.sw.Remove(w)
	// close the channel used to send messages to the client
	close(w.msgs)
	// close the error channel
	close(w.err)
	log.Println("client stopped", err)
}

var _ stopwatch.Observer = (*StopWatchWSClient)(nil)
