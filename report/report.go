package report

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/oaktown/calliope/store"
  "html/template"
  "io"
  "log"
)

type HtmlReport struct {
  Html  template.HTML
  Query string
}

type QueryOptions struct {
  StartDate     string
  EndDate       string
  Timezone      string
  Participants  string
  BodyOrSubject string
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
  Messages  []*store.Message
}

type JsonReport struct {
  Query     string
  ChartData Chart
  Messages  []*store.Message
}

func GetJsonReport(opt QueryOptions, svc *store.Service) JsonReport {
  search := setupMessageSearch(opt, svc)
  messages, _ := search.Do()
  chartData := getChartData(messages)

  return JsonReport{
    Query:     search.QueryString(),
    ChartData: chartData,
    Messages:  messages,
  }
}

func GetHtmlReport(opt QueryOptions, svc *store.Service) HtmlReport {
  messageSearch := setupMessageSearch(opt, svc)

  var reportBuffer bytes.Buffer
  Render(&reportBuffer, messageSearch, opt.InboxUrl)
  reportHtml := template.HTML(reportBuffer.String())

  return HtmlReport{
    Query: messageSearch.QueryString(),
    Html:  reportHtml,
  }
}

func setupMessageSearch(opt QueryOptions, svc *store.Service) store.MessageSearch {
  var messageSearch store.MessageSearch
  if opt.Query != "" {
    messageSearch = svc.NewRawMessageSearch(opt.Query)
  } else {
    messageSearch = svc.NewStructuredMessageSearch().
      Label(opt.Label).
      DateRange(opt.StartDate, opt.EndDate, opt.Timezone).
      Participants(opt.Participants).
      BodyOrSubject(opt.BodyOrSubject).
      Size(opt.Size).
      Sort(opt.SortField, opt.SortAscending).
      Starred(opt.Starred)
  }
  return messageSearch
}

func Render(wr io.Writer, messageSearch store.MessageSearch, inboxUrl string) {
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

func getChartData(messages []*store.Message) Chart {
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
