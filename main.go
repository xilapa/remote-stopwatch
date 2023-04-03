package main

import (
	"net/http"
	"regexp"
	"text/template"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	sw "github.com/xilapa/remote-stopwatch/stopwatch"
	"github.com/xilapa/remote-stopwatch/wsobserver"
	"nhooyr.io/websocket"
)

var (
	templates   = template.Must(template.ParseFiles("pages/home.html", "pages/stopwatch.html"))
	stopwatchs  = cmap.New[*sw.StopWatch]()
	idValidator = regexp.MustCompile("^/(join|syncwatch)/([A-Za-z0-9_-]+)$")
)

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/create", create)
	http.HandleFunc("/join/", join)
	http.HandleFunc("/syncwatch/", syncwatch)
	http.ListenAndServe(":8080", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", stopwatchs.Count())
}

func create(w http.ResponseWriter, r *http.Request) {
	stopwatch := sw.NewStopWatch()
	stopwatchs.Set(stopwatch.Id, stopwatch)
	http.Redirect(w, r, "/join/"+stopwatch.Id, http.StatusFound)
}

type observerView struct {
	ObserversCount int
	Id             string
	CurrentTime    time.Duration
}

func join(w http.ResponseWriter, r *http.Request) {
	stopwatch := getStopwatchFromPath(r.URL.Path)
	if stopwatch == nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	view := observerView{
		ObserversCount: stopwatch.ObserversCount(),
		Id:             stopwatch.Id,
		CurrentTime:    stopwatch.CurrentTime,
	}

	templates.ExecuteTemplate(w, "stopwatch.html", view)
}

func syncwatch(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	stopwatch := getStopwatchFromPath(r.URL.Path)

	// just for testing
	stopwatch.Start()
	if stopwatch == nil {
		c.Close(websocket.StatusNormalClosure, "stopwatch not found")
		return
	}

	obs := wsobserver.NewWebSocketObserver(r.Context(), c)
	stopwatch.Add(obs)
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
