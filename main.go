package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"text/template"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	sw "github.com/xilapa/remote-stopwatch/stopwatch"
	swclient "github.com/xilapa/remote-stopwatch/stopwatchclient"
	"nhooyr.io/websocket"
)

var (
	templates   = template.Must(template.ParseFiles("pages/home.html", "pages/stopwatch.html"))
	stopwatchs  = cmap.New[*sw.StopWatch]()
	idValidator = regexp.MustCompile("^/(join|syncwatch)/([A-Za-z0-9_-]+)$")
	fiveMinutes = 5 * time.Minute
)

func main() {
	cfg, err := NewConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", home)
	http.HandleFunc("/create", create)
	http.HandleFunc("/join/", join)
	http.HandleFunc("/syncwatch/", syncwatch)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go removeIdleStopwatchs(sigs)

	http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", stopwatchs.Count())
}

func create(w http.ResponseWriter, r *http.Request) {
	stopwatch := sw.NewStopWatch(sw.WithTimeLoopInterval(250 * time.Millisecond))
	stopwatchs.Set(stopwatch.Id, stopwatch)
	http.Redirect(w, r, "/join/"+stopwatch.Id, http.StatusFound)
}

func join(w http.ResponseWriter, r *http.Request) {
	stopwatch := getStopwatchFromPath(r.URL.Path)
	if stopwatch == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	view := stopwatch.GetCurrentView()
	templates.ExecuteTemplate(w, "stopwatch.html", view)
}

func syncwatch(w http.ResponseWriter, r *http.Request) {
	stopwatch := getStopwatchFromPath(r.URL.Path)

	if stopwatch == nil {
		return
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	obs := swclient.NewWebSocketClient(r.Context(), c)
	obs.Handle(stopwatch)
}

func getStopwatchFromPath(path string) *sw.StopWatch {
	idMatch := idValidator.FindStringSubmatch(path)
	if idMatch == nil || len(idMatch) != 3 {
		return nil
	}
	stopWatch, ok := stopwatchs.Get(idMatch[2])
	if !ok {
		return nil
	}
	return stopWatch
}

func removeIdleStopwatchs(signals <-chan os.Signal) {
	for {
		select {
		case <-time.After(fiveMinutes):
			keysToRemove := make([]string, 0, stopwatchs.Count())

			stopwatchs.IterCb(func(key string, sw *sw.StopWatch) {
				if (sw.IdleSince != time.Time{} && time.Since(sw.IdleSince) > fiveMinutes) {
					keysToRemove = append(keysToRemove, key)
				}
			})

			for i := range keysToRemove {
				stopwatchs.Remove(keysToRemove[i])
			}

		case sig := <-signals:
			fmt.Println("Terminating... ", sig)
			return
		}
	}
}
