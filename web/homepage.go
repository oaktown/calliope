package web

import (
  "github.com/oaktown/calliope/report"
  "github.com/oaktown/calliope/store"
  "html/template"
  "log"
  "net/http"
  "time"
)

type Data struct {
  Title       string
  Fields      report.QueryOptions
  Size        int
  Query       string
  Report      template.HTML
  Earliest    time.Time
  Latest      time.Time
  TotalEmails int64
}

func ShowHomePage(client *store.Service, w http.ResponseWriter, opt report.QueryOptions) {
  rpt := report.GetReport(opt, client)

  // TODO: maybe move this to another page and add a link
  stats, _ := client.GetStats()
  data := Data{
    Title:       "Calliope Email Report",
    Fields:      opt,
    Query:       rpt.Query,
    Report:      rpt.Html,
    Earliest:    stats.Earliest,
    Latest:      stats.Latest,
    TotalEmails: stats.Total,
  }

  t := template.Must(template.ParseFiles("templates/layout.html", "templates/web-ui.html"))

  if err := t.ExecuteTemplate(w, "layout", data); err != nil {
    log.Println("Error occurred while executing template: ", err)
  }
}
