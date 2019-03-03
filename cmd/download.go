package cmd

import (
	"fmt"
	"github.com/oaktown/calliope/gmailservice"
	"github.com/oaktown/calliope/misc"
	"github.com/oaktown/calliope/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"time"
)

var limit, query, inboxUrl string

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&limit, "limit", "l", "10", "limit number of emails to download (if > 500, rounds up to next multiple of 500).")
	downloadCmd.Flags().StringVarP(&query, "query", "q", "", "download based on Gmail query. E.g. \"after: 2018/11/01 label:my-label is:starred\" More info: See https://support.google.com/mail/answer/7190.")
	downloadCmd.Flags().StringVarP(&inboxUrl, "inbox-url", "u", "https://mail.google.com/mail/", "Url for gmail (useful if you are logged into multiple accounts).")
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "downloads emails",
	Long:  `Downloads emails from Gmail into Elasticsearch. Reports can then be run on the emails that have been downloaded.`,
	Run: func(cmd *cobra.Command, args []string) {
		download()
	},
}

func reader(s *store.Service, messageChannel <-chan *store.Message, maxWorkers int) {
	workers := make(chan bool, maxWorkers)
	duplicates := make(chan *store.MessageResponse, 100000)
	var savedMessages, errors int64
	for message := range messageChannel { // reads from channel until it's closed
		workers <- true
		go func() {
			defer func() { <-workers }()
			// TODO: Determine if ids are ever duplicates. Although it's weird …the number of messages total was different on two different runs of 100k. One was 94224, the other was 94092
			// TODO: add another channel for verify
			err := s.SaveMessage(*message, duplicates)
			if err != nil {
				log.Printf("Error saving id %s: %s\n", message.Id, err)
				errors++
			} else {
				log.Printf("Saved id %s: %s\n", message.Id, message.Subject)
				savedMessages++
			}
		}()
	}
	for i := 0; i < maxWorkers; i++ {
		workers <- true
	}
	fmt.Println("Total messages:", savedMessages+errors)
	fmt.Println("Total saved messages: ", savedMessages)
	fmt.Println("Total errors: ", errors)
	fmt.Println("Total duplicates: ", len(duplicates))
	fmt.Println("Net messages: ", savedMessages-int64(len(duplicates)))
}

func download() {
	startedAt := time.Now()
	fmt.Println("Started at:", startedAt)
	excludeHeaders := viper.GetStringMapStringSlice("exclude_headers_with_values")
	for k, v := range excludeHeaders {
		fmt.Printf("key: %s\n", k)
		for _, s := range v {
			fmt.Printf("  %s\n", s)
		}
	}
	max, _ := strconv.ParseInt(limit, 10, 64)

	gsvc := misc.GetGmailClient()
	s := misc.GetStoreClient()
	options := gmailservice.Options{
		Query:          query,
		Limit:          max,
		InboxUrl:       inboxUrl,
		ExcludeHeaders: excludeHeaders,
	}
	d := gmailservice.New(gsvc, options, 200)
	labels := gmailservice.Download(d)
	if err := s.SaveLabels(labels); err != nil {
		log.Println("Error saving labels")
	}

	reader(s, d.MessageChan, 10)
	finishedAt := time.Now()
	fmt.Println("Started at:", startedAt)
	fmt.Println("Time ended", finishedAt)
	fmt.Printf("Elapsed time: %f seconds\n", finishedAt.Sub(startedAt).Seconds())
}
