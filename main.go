package main

import (
	"net/http"
	"text/template"

	cmap "github.com/orcaman/concurrent-map/v2"
	sw "github.com/xilapa/remote-stopwatch/stopwatch"
)

var (
	templates  = template.Must(template.ParseFiles("pages/home.html"))
	stopwatchs = cmap.New[sw.StopWatch]()
)

func main() {

	// just for testing
	stopwatchs.Set("1", *sw.NewStopWatch())

	http.HandleFunc("/", home)
	http.ListenAndServe(":8080", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", stopwatchs.Count())
}
