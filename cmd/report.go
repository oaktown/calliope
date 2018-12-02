package cmd

import (
  "github.com/spf13/cobra"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
)

var label string

func init() {
  rootCmd.AddCommand(reportCmd)
  reportCmd.Flags().StringVarP(&label, "label", "l", "",
    "Report for emails with this label which are also starred (required).")
  reportCmd.MarkFlagRequired("label")
}

var reportCmd = &cobra.Command{
  Use:   "report",
  Short: "Creates an HTML report of labeled emails",
  Long: `Creates an HTML report of emails with the 
specified label which are also starred (because the 
threaded UI in Gmail only allows applying labels to 
threads).`,
  Run: func(cmd *cobra.Command, args []string) {
    report.Run(misc.GetGmailClient(), label)
  },
}
