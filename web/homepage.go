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
  Fields      Options
  Size        int
  Query       string
  Report      template.HTML
  Earliest    time.Time
  Latest      time.Time
  TotalEmails int64
}

//// Create your search request
//ss := elastic.NewSearchSource().Query(elastic.NewMatchAllQuery()).From(0).Size(10)
//data, _ := json.Marshal(ss.Source())
//fmt.Printf("%s", string(data))
//...
//// Use ss in search
//res, err := client.Search().SearchSource(ss).Do()
//...
type Options struct {
  Label         string
  InboxUrl      string
  Size          int
  Starred       bool
  SortField     string
  SortAscending bool
  Query         string
}

func ShowHomePage(client *store.Service, query string, w http.ResponseWriter, opt Options) {
  stats, _ := client.GetStats()
  reportHtml := template.HTML(getReportHtml(client, query, opt.Size, opt.InboxUrl))
  data := Data{
    Title:       "Calliope Email Report",
    Fields:      opt,
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

//
//func (q *Query) FromLabel(client *store.Service, opt Options) *Query {
//  query, _ := client.GetQueryFromLabel(opt.Label, opt.Starred, opt.Size)
//  q.Q = query
//return q
//}

func QueryStringFromLabel(client *store.Service, opt Options) string {
  if opt.Label == "" {
    log.Println("No label")
    return ""
  }
  boolQuery, _ := client.GetQueryFromLabel(opt.Label, opt.Starred, opt.Size)
  source, _ := boolQuery.Source()
  q, _ := json.MarshalIndent(source, "  ", "  ")
  query := fmt.Sprintf("{\n  \"query\": %s\n}", q)
  log.Println("Query:", query)
  return query
}
