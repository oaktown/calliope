package web

import (
  "github.com/oaktown/calliope/misc"
  "html/template"
  "log"
  "net/http"
  "time"
)

type Stats struct {
  Title       string
  Earliest    time.Time
  Latest      time.Time
  TotalEmails int64
}

func ShowStats(r *http.Request, w http.ResponseWriter) {
  stats, _ := misc.GetStoreClient().GetStats()
  data := Stats{
    Title: "server stats",
    Earliest: stats.Earliest,
    Latest:      stats.Latest,
    TotalEmails: stats.Total,
  }

  t := template.Must(template.ParseFiles("templates/layout.html", "templates/stats.html"))

  if err := t.ExecuteTemplate(w, "layout", data); err != nil {
    log.Println("Error occurred while executing status template: ", err)
  }
}
