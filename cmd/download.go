package cmd

import (
  "log"
  "strconv"
  "sync"

  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/store"
  "github.com/spf13/cobra"
)

var limit, lastDate, pageToken, inboxUrl string

func init() {
  rootCmd.AddCommand(downloadCmd)
  downloadCmd.Flags().StringVarP(&limit, "limit", "l", "10", "limit number of emails to download (if > 500, rounds up to next multiple of 500).")
  downloadCmd.Flags().StringVarP(&lastDate, "after-date", "d", "", "Emails after this date. In yyyy/mm/dd format.")
  downloadCmd.Flags().StringVarP(&pageToken, "page-token", "p", "", "Page token for downloading emails (probably going to be removed).")
  downloadCmd.Flags().StringVarP(&inboxUrl, "inbox-url", "u", "https://mail.google.com/mail/#inbox/", "Url for gmail (useful if you are logged into multiple accounts).")
}

var downloadCmd = &cobra.Command{
  Use:   "download",
  Short: "downloads emails",
  Long:  `Downloads emails into Elasticsearch.`,
  Run: func(cmd *cobra.Command, args []string) {
    download()
  },
}

func reader(s store.Storable, messageChannel <-chan *gmailservice.Message, wg *sync.WaitGroup) {
  defer wg.Done() // WaitGroup done when this routines exits

  for message := range messageChannel { // reads from channel until it's closed
    s.Save(*message)
  }
}

func download() {
  max, _ := strconv.ParseInt(limit, 10, 64)
  gsvc := misc.GetGmailClient()
  s := misc.GetStoreClient()
  var wg sync.WaitGroup
  const BufferSize = 10
  storageChannel := make(chan *gmailservice.Message, BufferSize)
  wg.Add(1)
  go reader(s, storageChannel, &wg)
  options := gmailservice.Options{
    LastDate: lastDate,
    Limit: max,
    InboxUrl: inboxUrl,
  }
  d := gmailservice.New(gsvc, options, 200)
  messages, err := gmailservice.Download(d)

  log.Print("Done downloading: ", len(messages))
  if err != nil {
   log.Fatal("Unable to download message. Error: ", err)
  }
  for _, message := range messages {
   storageChannel <- message
  }
  close(storageChannel)
  wg.Wait()
}
