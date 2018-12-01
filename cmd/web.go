package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

var port int

func init() {
  rootCmd.AddCommand(webCmd)
  webCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run web server on.")
}

var webCmd = &cobra.Command{
  Use:   "web",
  Short: "Start web server",
  Long:  `Starts up the Calliope web app.`,
  Run: func(cmd *cobra.Command, args []string) {
    web()
  },
}

func web() {
  fmt.Println("TBD")
}
