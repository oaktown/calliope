package report

import (
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
  ChartJson []BarData
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

  chartJson := getChartData(messages)

  report := template.Must(
    template.New("report.html").
      Funcs(template.FuncMap{
        "gmailUrl": gmailUrl,
        "jump":     jump,
      }).
      ParseFiles("templates/report.html"))
  data := Data{
    ChartJson: chartJson,
    Messages:  messages,
  }
  if err := report.Execute(wr, data); err != nil {
    log.Println("Error rendering template: ", err)
  }
}

func getChartData(messages []*gmailservice.Message) []BarData {
  data := make(map[string]int)
  for _, m := range messages {
    date := m.Date.Format("2006-01-02")
    data[date] = data[date] + 1
  }
  var chart []BarData
  for d, n := range data {
    chart = append(chart, BarData{
      Date:     d,
      Messages: n,
    })
  }

  return chart
}
