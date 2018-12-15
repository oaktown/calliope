package web

import (
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
  "github.com/oaktown/calliope/store"
  "html/template"
  "log"
  "net/http"
  "strconv"
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

func ShowHomepage(r *http.Request, w http.ResponseWriter) {
  log.Println("Request path:", r.URL.Path)
  client := misc.GetStoreClient()
  var sortField string
  if sortField = r.FormValue("sort"); sortField == "" {
    sortField = "Date"
  }
  inboxUrl := r.FormValue("gmailurl")
  if inboxUrl == "" {
    inboxUrl = "https://mail.google.com/mail/"
  }
  size, err := strconv.Atoi(r.FormValue("size"))
  if err != nil {
    size = 100
  }
  var timezone string
  if timezone = r.FormValue("timezone"); timezone == "" {
    timezone = "-0800" // Default to PST
  }
  opt := report.QueryOptions{
    StartDate:     r.FormValue("startDate"),
    EndDate:       r.FormValue("endDate"),
    Timezone:      timezone,
    Participants:  r.FormValue("participants"),
    Label:         r.FormValue("label"),
    Starred:       r.FormValue("starred") == "true",
    InboxUrl:      inboxUrl,
    Size:          size,
    SortField:     sortField,
    SortAscending: r.FormValue("ascending") == "true",
    Query:         r.FormValue("query"),
  }
  log.Printf("options: %+v\n", opt)
  renderHomepage(client, w, opt)
}

func renderHomepage(client *store.Service, w http.ResponseWriter, opt report.QueryOptions) {
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
