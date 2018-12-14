package report

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
  "github.com/olivere/elastic"
  "html/template"
  "io"
  "log"
)

type QueryOptions struct {
  StartDate     string
  EndDate       string
  Participants  string
  Label         string
  InboxUrl      string
  Size          int
  Starred       bool
  SortField     string
  SortAscending bool
  Query         string
}

type Options struct {
  Label    string
  Starred  bool
  InboxUrl string
  Size     int
}

type BarData struct {
  Date     string
  Messages int
}

type Data struct {
  ChartJson template.HTML
  Messages  []*gmailservice.Message
}

//report.Run(client, &reportBuffer, req, inboxUrl)
func Run(s *store.Service, wr io.Writer, req *elastic.SearchService, inboxUrl string) {

  gmailUrl := func(threadId string) string {
    return fmt.Sprintf("%v#inbox/%v", inboxUrl, threadId)
  }

  jump := func(id string) string {
    return fmt.Sprint("#", id)
  }

  messages, err := s.GetMessages(req)
  if err != nil {
    log.Println("Exiting due to error")
    return
  }

  chartData := getChartData(messages)
  chartJson, _ := json.MarshalIndent(chartData, "", "  ")

  report := template.Must(
    template.New("report.html").
      Funcs(template.FuncMap{
        "gmailUrl": gmailUrl,
        "jump":     jump,
      }).
      ParseFiles("templates/report.html"))
  data := Data{
    ChartJson: template.HTML(chartJson),
    Messages:  messages,
  }
  if err := report.Execute(wr, data); err != nil {
    log.Println("Error rendering template: ", err)
  }
}

func getChartData(messages []*gmailservice.Message) []BarData {
  if len(messages) == 0 {
    return nil
  }
  data := make(map[string]int)
  first := messages[0].Date
  last := messages[0].Date
  for _, m := range messages {
    date := m.Date

    if date.After(last) {
      last = date
    }
    if date.Before(first) {
      first = date
    }

    dateString := date.Format("2006-01-02")
    data[dateString] = data[dateString] + 1
  }
  chart := make([]BarData, 0, len(messages))
  for d := first; d.Before(last.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
    dateString := d.Format("2006-01-02")
    chart = append(chart, BarData{
      Date:     dateString,
      Messages: data[dateString],
    })

  }
  return chart
}

type Report struct {
  Html template.HTML
  Query string
  ChartData string // Maybe for later; right now this can be part of html
}

func GetReport(opt QueryOptions, client *store.Service) (Report) {
  var query string
  if opt.Query == "" {
    query = QueryStringFromLabel(client, opt)
  } else {
    query = opt.Query
  }

  reportHtml := template.HTML(GetReportHtml(client, query, opt.Size, opt.InboxUrl))
  return Report{
    Query: query,
    Html: reportHtml,
  }
}

func GetReportHtml(client *store.Service, query string, size int, inboxUrl string) string {
  // TODO: Change this to return html (which is limited to size), the chart data which should be everything within
  // reason, and the query JSON.
  // Also, move this to report.
  var req *elastic.SearchService
  req = client.GetRawQuery(query, size)
  var reportBuffer bytes.Buffer
  Run(client, &reportBuffer, req, inboxUrl)
  return reportBuffer.String()
}

func QueryStringFromLabel(client *store.Service, opt QueryOptions) string {
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
