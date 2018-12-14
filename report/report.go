package report

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
  "html/template"
  "io"
  "log"
)

type Report struct {
  Html  template.HTML
  Query string
}

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

type BarData struct {
  Date     string
  Messages int
}

type Chart []BarData

type HtmlData struct {
  ChartJson template.HTML
  Messages  []*gmailservice.Message
}

type ReportGenerator struct {
  Opt QueryOptions
  Svc *store.Service
}

func GetReport(opt QueryOptions, svc *store.Service) Report {
  generator := ReportGenerator{
    Opt: opt,
    Svc: svc,
  }
  var messageSearch store.MessageSearch
  if opt.Query != "" {
    messageSearch = svc.NewRawMessageSearch(opt.Query)
  } else {
    messageSearch = svc.NewStructuredMessageSearch().Label(opt.Label).StartDate(opt.StartDate).EndDate(opt.EndDate).Participants(opt.EndDate).Size(opt.Size)
  }
  return Report{
    Query: messageSearch.QueryString(),
    Html:  generator.GetReportHtml(messageSearch),
  }
}

func (r ReportGenerator) GetReportHtml(messageSearch store.MessageSearch) template.HTML {
  var reportBuffer bytes.Buffer
  Render(r.Svc, &reportBuffer, messageSearch, r.Opt.InboxUrl)
  return template.HTML(reportBuffer.String())
}

func Render(s *store.Service, wr io.Writer, messageSearch store.MessageSearch, inboxUrl string) {
  gmailUrl := func(threadId string) string {
    return fmt.Sprintf("%v#inbox/%v", inboxUrl, threadId)
  }

  jump := func(id string) string {
    return fmt.Sprint("#", id)
  }

  // TODO: Return error or HTML that indicates a problem
  messages, err := messageSearch.Do()
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
  data := HtmlData{
    ChartJson: template.HTML(chartJson),
    Messages:  messages,
  }
  if err := report.Execute(wr, data); err != nil {
    log.Println("Error rendering template: ", err)
  }
}

func getChartData(messages []*gmailservice.Message) Chart {
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
