package main

import (
	"net/http"
	"text/template"

	cmap "github.com/orcaman/concurrent-map/v2"
	sw "github.com/xilapa/remote-stopwatch/stopwatch"
)

var (
	templates  = template.Must(template.ParseFiles("pages/home.html"))
	stopwatchs = cmap.New[*sw.StopWatch]()
)

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/create", create)
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
