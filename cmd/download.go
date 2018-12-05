package cmd

import (
  "fmt"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/store"
  "github.com/spf13/cobra"
  "html/template"
  "log"
  "net/url"
  "os"
  "strconv"
)

var limit, query, pageToken, inboxUrl string
var runReport bool

func init() {
  rootCmd.AddCommand(downloadCmd)
  downloadCmd.Flags().StringVarP(&limit, "limit", "l", "10", "limit number of emails to download (if > 500, rounds up to next multiple of 500).")
  downloadCmd.Flags().StringVarP(&query, "query", "q", "", "Gmail query. E.g. \"after: 2018/11/01 label:my-label is:starred\" More info: See https://support.google.com/mail/answer/7190.")
  downloadCmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "Page token for downloading emails (probably going to be removed).")
  downloadCmd.Flags().StringVarP(&inboxUrl, "inbox-url", "u", "https://mail.google.com/mail/", "Url for gmail (useful if you are logged into multiple accounts).")
  downloadCmd.Flags().BoolVarP(&runReport, "run-report", "R", false, "Runs a report instead of saving to Elasticsearch (in the future, this will be a different command altogether)")
}

var downloadCmd = &cobra.Command{
  Use:   "download",
  Short: "downloads emails",
  Long:  `Downloads emails into Elasticsearch.`,
  Run: func(cmd *cobra.Command, args []string) {
    download()
  },
}

func reader(s store.Storable, messageChannel <-chan *gmailservice.Message, workers chan bool) {
  for message := range messageChannel { // reads from channel until it's closed
    workers <- true
    go func() {
      defer func() { <-workers }()
      err := s.Save(*message)
      if err != nil {
        log.Println("Error saving: ", err)
      } else {
        log.Println("Saved:\n  ", message.Subject)
      }
    }()
  }
}

type ReportData struct {
  Query    string
  Messages []*gmailservice.Message
}

func gmailUrl(threadId string) string {
  return fmt.Sprintf("%v#search/%v/%v", inboxUrl, url.QueryEscape(query), threadId)
}

func jump(id string) string {
  return fmt.Sprint("#", id)
}

func generateReport(s store.Storable, ch <-chan *gmailservice.Message) {

  var messages []*gmailservice.Message
  for msg := range ch {
    messages = append(messages, msg)
  }
  log.Println("Messages found: ", len(messages))

  report := template.Must(
    template.New("report.html").
      Funcs(template.FuncMap{
        "gmailUrl": gmailUrl,
        "jump":     jump,
      }).
      ParseFiles("report.html"))
  data := ReportData{query, messages}
  err := report.Execute(os.Stdout, data)
  if err != nil {
    log.Println("Error rendering template: ", err)
  }
}

func download() {
  max, _ := strconv.ParseInt(limit, 10, 64)
  maxWorkers := 10
  workers := make(chan bool, maxWorkers)

  gsvc := misc.GetGmailClient()
  s := misc.GetStoreClient()
  options := gmailservice.Options{
    Query:    query,
    Limit:    max,
    InboxUrl: inboxUrl,
  }
  d := gmailservice.New(gsvc, options, 200)
  gmailservice.Download(d)
  if runReport {
    generateReport(s, d.MessageChan)
  } else {
    reader(s, d.MessageChan, workers)
  }

  for i := 0; i < maxWorkers; i++ {
    workers <- true
  }
}
