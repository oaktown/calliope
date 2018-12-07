package cmd

import (
  "github.com/spf13/cobra"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
)

var label, url string
var allMessages bool
var size int

func init() {
  rootCmd.AddCommand(reportCmd)
  reportCmd.Flags().StringVarP(&label, "label", "l", "",
    "Report for emails with this label (required).")
  reportCmd.MarkFlagRequired("label")
  reportCmd.Flags().BoolVarP(&allMessages, "all-messages", "A", false, "By default, we only query for starred messages. With this flag, we get all messages for the label whether they are starred or not.")
  reportCmd.Flags().IntVarP(&size, "size", "s", 1000, "The max number of results to return from a search. Defaults to 1000.")
  downloadCmd.Flags().StringVarP(&url, "url", "U", "https://mail.google.com/mail/", "Url for gmail (useful if you are logged into multiple accounts).")
}

var reportCmd = &cobra.Command{
  Use:   "report",
  Short: "Creates an HTML report of labeled emails",
  Long: `Creates an HTML report of emails with the specified label which are also starred (because the threaded UI in Gmail only allows applying labels to threads). Report queries emails saved to Elasticsearch (as opposed to doing a search on Gmail).`,
  Run: func(cmd *cobra.Command, args []string) {
    options := report.Options{
      Label: label,
      Starred: !allMessages,
      InboxUrl: url,
      Size: size,
    }
    report.Run(misc.GetStoreClient(), options)
  },
}
