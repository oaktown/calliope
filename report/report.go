package report

import (
  "encoding/json"
  "fmt"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
  "github.com/olivere/elastic"
  "html/template"
  "io"
  "log"
)

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
