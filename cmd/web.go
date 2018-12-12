package cmd

import (
  "fmt"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/web"
  "github.com/spf13/cobra"
  "log"
  "net/http"
)

var port string
var allMessages bool
var options = web.Options{}

func init() {
  rootCmd.AddCommand(webCmd)
  webCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run web server on.")
  webCmd.Flags().StringVarP(&options.Label, "label", "l", "",
    "Report for emails with this label (required).")
  webCmd.MarkFlagRequired("label")
  webCmd.Flags().BoolVarP(&allMessages, "all-messages", "A", false, "By default, we only query for starred messages. With this flag, we get all messages for the label whether they are starred or not.")
  webCmd.Flags().IntVarP(&options.Size, "size", "s", 1000, "The max number of results to return from a search. Defaults to 1000.")
  webCmd.Flags().StringVarP(&options.InboxUrl, "url", "U", "https://mail.google.com/mail/", "Url for gmail (useful if you are logged into multiple accounts).")
}

var webCmd = &cobra.Command{
  Use:   "web",
  Short: "Start web server",
  Long:  `Starts up the Calliope web app.`,
  Run: func(cmd *cobra.Command, args []string) {
    options.Starred = !allMessages
    startServer(options)
  },
}

func startServer(opt web.Options) {
  client := misc.GetStoreClient()
  initialQuery := true

  http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "images/email_monster_Uxt_icon.ico")
  })

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    var query string

    if initialQuery {
      query = web.QueryStringFromLabel(client, opt)
      initialQuery = false
    } else {
      query = r.FormValue("query")
    }

    web.ShowHomePage(client, query, w, opt)
  })

  fmt.Printf("Starting web server: http://localhost:%s/\n\n", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}
