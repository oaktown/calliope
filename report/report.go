package report

import (
  "encoding/base64"
  "github.com/microcosm-cc/bluemonday"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
  "html/template"
)

type MessageWithHtml struct {
  store.Message
  BodyHtml template.HTML
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

type JsonReport struct {
  Query     string
  ChartData Chart
  Messages  []*MessageWithHtml
}

func GetJsonReport(opt QueryOptions, svc *store.Service) JsonReport {
  search := setupMessageSearch(opt, svc)
  messages, _ := search.Do()
  reportMessages := FillInHtmlBody(messages)

  chartData := getChartData(messages)

  return JsonReport{
    Query:     search.QueryString(),
    ChartData: chartData,
    Messages:  reportMessages,
  }
}

func FillInHtmlBody(messages []*store.Message) []*MessageWithHtml {
  var messagesWithHtml []*MessageWithHtml
  for _, message := range messages {
    m := &MessageWithHtml{Message: *message, BodyHtml : template.HTML(GetMessageHtmlBody(*message))}
    messagesWithHtml = append(messagesWithHtml, m)
  }
  return messagesWithHtml
}

func GetMessageHtmlBody(message store.Message) string {
  htmlBodyEncoded := gmailservice.GetBodyPartByMimeType(message.Source, "text/html")
  htmlBody, _ := base64.URLEncoding.DecodeString(htmlBodyEncoded)
  var unsafeHtml string
  if len(htmlBody) > 0 {
    unsafeHtml = string(htmlBody)
  } else {
    unsafeHtml = "<pre>\n" + message.Body + "\n</pre>"
  }

  p := bluemonday.UGCPolicy()
  html := p.Sanitize(unsafeHtml)
  return html
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
