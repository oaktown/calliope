package cmd

import (
  "github.com/spf13/cobra"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
)

var label, esQuery string
var allMessages bool

func init() {
  rootCmd.AddCommand(reportCmd)
  //reportCmd.Flags().StringVarP(&esQuery, "query", "q", "", "Elasticsearch query (overrides other flags).")
  reportCmd.Flags().StringVarP(&label, "label", "l", "",
    "Report for emails with this label (required).")
  reportCmd.MarkFlagRequired("label")
  reportCmd.Flags().BoolVarP(&allMessages, "all-messages", "A", false, "By default, we only query for starred messages. With this flag, we get all messages for the label whether they are starred or not.")
}

var reportCmd = &cobra.Command{
  Use:   "report",
  Short: "Creates an HTML report of labeled emails",
  Long: `Creates an HTML report of emails with the specified label which are also starred (because the threaded UI in Gmail only allows applying labels to threads). Report queries emails saved to Elasticsearch (as opposed to doing a search on Gmail).`,
  Run: func(cmd *cobra.Command, args []string) {
    options := report.Options{
      Label: label,
      Starred: !allMessages,
      Query: esQuery,
    }
    report.Run(misc.GetStoreClient(), options)
  },
}
