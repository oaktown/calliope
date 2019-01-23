package web

import (
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
  "github.com/oaktown/calliope/store"
  "html/template"
  "log"
  "net/http"
)

type HomepageData struct {
  Title       string
  Fields      report.QueryOptions
  Size        int
  Query       string
  Report      template.HTML
}

type Homepage struct {
  r   *http.Request
  w   http.ResponseWriter
  svc *store.Service
}

func ShowOldHomePage(r *http.Request, w http.ResponseWriter, opts report.QueryOptions) {
  h := Homepage{
    r:   r,
    w:   w,
    svc: misc.GetStoreClient(),
  }
  h.render(opts)
}

func (h Homepage) render(opt report.QueryOptions) {
  rpt := report.GetHtmlReport(opt, h.svc)

  data := HomepageData{
    Title:       "Calliope Email Report",
    Fields:      opt,
    Query:       rpt.Query,
    Report:      rpt.Html,
  }

  t := template.Must(template.ParseFiles("templates/layout.html", "templates/old_homepage.html"))

  if err := t.ExecuteTemplate(h.w, "layout", data); err != nil {
    log.Println("Error occurred while executing homepage template: ", err)
  }
}
