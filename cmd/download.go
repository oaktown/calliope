package cmd

import (
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/store"
  "github.com/spf13/cobra"
  "log"
  "strconv"
)

var limit, query, inboxUrl string
var runReport bool

func init() {
  rootCmd.AddCommand(downloadCmd)
  downloadCmd.Flags().StringVarP(&limit, "limit", "l", "10", "limit number of emails to download (if > 500, rounds up to next multiple of 500).")
  downloadCmd.Flags().StringVarP(&query, "query", "q", "", "Gmail query. E.g. \"after: 2018/11/01 label:my-label is:starred\" More info: See https://support.google.com/mail/answer/7190.")
  downloadCmd.Flags().StringVarP(&inboxUrl, "inbox-url", "u", "https://mail.google.com/mail/", "Url for gmail (useful if you are logged into multiple accounts).")
}

var downloadCmd = &cobra.Command{
  Use:   "download",
  Short: "downloads emails",
  Long:  `Downloads emails into Elasticsearch.`,
  Run: func(cmd *cobra.Command, args []string) {
    download()
  },
}

func reader(s *store.Service, messageChannel <-chan *gmailservice.Message, workers chan bool) {
  for message := range messageChannel { // reads from channel until it's closed
    workers <- true
    go func() {
      defer func() { <-workers }()
      err := s.SaveMessage(*message)
      if err != nil {
        log.Println("Error saving: ", err)
      } else {
        log.Println("Saved:\n  ", message.Subject)
      }
    }()
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
  labels := gmailservice.Download(d)
  if err := s.SaveLabels(labels); err != nil {
    log.Println("Error saving labels")
  }

  reader(s, d.MessageChan, workers)

  for i := 0; i < maxWorkers; i++ {
    workers <- true
  }
}
