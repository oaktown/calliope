package cmd

import (
  "fmt"
  "github.com/oaktown/calliope/auth"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
  "github.com/spf13/cobra"
  "golang.org/x/net/context"
  "google.golang.org/api/gmail/v1"
  "log"
  "sync"
)

var limit int
var lastDate, pageToken, inboxUrl string

func init() {
  rootCmd.AddCommand(downloadCmd)
  downloadCmd.Flags().IntVarP(&limit, "limit", "l", 10, "limit number of emails to download")
  downloadCmd.Flags().StringVarP(&lastDate, "after-date", "d", "2018/01/01", "Emails after this date. In yyyy/mm/dd format.")
  downloadCmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "Page token for downloading emails (probably going to be removed).")
  downloadCmd.Flags().StringVarP(&inboxUrl, "inbox-url", "u", "https://mail.google.com/mail/#inbox/", "Url for gmail (useful if you are logged into multiple accounts).")
}

var downloadCmd = &cobra.Command{
  Use:   "download",
  Short: "downloads emails",
  Long:  `Downloads emails into Elasticsearch.`,
  Run: func(cmd *cobra.Command, args []string) {
    //fmt.Println("Hello!")
    download()
  },
}

func reader(s store.Storable, messageChannel <-chan gmailservice.Message, wg *sync.WaitGroup) {
  defer wg.Done() // WaitGroup done when this routines exits

  for message := range messageChannel { // reads from channel until it's closed
    s.Save(message)
  }
}

func download() {

  ctx := context.Background()
  client, err := auth.Client(ctx)
  if err != nil {
    log.Fatalf("could not get auth client, %v", err)
  }
  gsvc, err := gmailservice.New(client)
  if err != nil {
    log.Fatalf("could not create gmailservice, %v", err)
  }

  s, err := store.New(ctx)
  if err != nil {
    log.Fatalf("could not create store, %v", err)
  }

  downloadGmailToES(s, err, gsvc)
}

func downloadGmailToES(s *store.Service, err error, gsvc *gmail.Service) {
  var wg sync.WaitGroup
  const BufferSize = 10
  messageChannel := make(chan gmailservice.Message, BufferSize)
  wg.Add(1)
  go reader(s, messageChannel, &wg)
  fmt.Print("url: ", inboxUrl)
  messages, err := gmailservice.Download(gsvc, lastDate, limit, pageToken, inboxUrl)
  if err != nil {
    log.Fatal("Unable to download messageChannel. Error: ", err)
  }
  for _, message := range messages {
    messageChannel <- message
  }
  close(messageChannel)
  wg.Wait()
}
