package web

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/oaktown/calliope/report"
  "github.com/oaktown/calliope/store"
  "github.com/olivere/elastic"
  "html/template"
  "log"
  "net/http"
  "time"
)

type Data struct {
  Title       string
  Size        int
  Query       string
  Report      template.HTML
  Earliest    time.Time
  Latest      time.Time
  TotalEmails int64
}

type Options struct {
  Label   string
  InboxUrl     string
  Size    int
  Starred bool
}

func ShowHomePage(client *store.Service, query string, w http.ResponseWriter, opt Options) {
  stats, _ := client.GetStats()
  reportHtml := template.HTML(getReportHtml(client, query, opt.Size, opt.InboxUrl))
  data := Data{
    Title:       "Calliope Email Report",
    Size:        opt.Size,
    Query:       query,
    Report:      reportHtml,
    Earliest:    stats.Earliest,
    Latest:      stats.Latest,
    TotalEmails: stats.Total,
  }

  t := template.Must(template.ParseFiles("templates/layout.html", "templates/web-ui.html"))

  if err := t.ExecuteTemplate(w, "layout", data); err != nil {
    log.Println("Error occurred while executing template: ", err)
  }
}

func getReportHtml(client *store.Service, query string, size int, inboxUrl string) string {
  var req *elastic.SearchService
  req = client.GetRawQuery(query, size)
  var reportBuffer bytes.Buffer
  report.Run(client, &reportBuffer, req, inboxUrl)
  return reportBuffer.String()
}

func QueryStringFromLabel(client *store.Service, opt Options) string {
  boolQuery, _ := client.GetQueryFromLabel(opt.Label, opt.Starred, opt.Size)
  source, _ := boolQuery.Source()
  q, _ := json.MarshalIndent(source, "  ", "  ")
  query := fmt.Sprintf("{\n  \"query\": %s\n}", q)
  fmt.Println("Query using label: ", query)
  return query
}
