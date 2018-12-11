package cmd

import (
  "bytes"
  "fmt"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
  "github.com/oaktown/calliope/store"
  "github.com/spf13/cobra"
  "html/template"
  "log"
  "net/http"
  "time"
)

var port string

func init() {
  rootCmd.AddCommand(webCmd)
  webCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run web server on.")
  webCmd.Flags().StringVarP(&label, "label", "l", "",
    "Report for emails with this label (required).")
  webCmd.MarkFlagRequired("label")
  webCmd.Flags().BoolVarP(&allMessages, "all-messages", "A", false, "By default, we only query for starred messages. With this flag, we get all messages for the label whether they are starred or not.")
  webCmd.Flags().IntVarP(&size, "size", "s", 1000, "The max number of results to return from a search. Defaults to 1000.")
  webCmd.Flags().StringVarP(&url, "url", "U", "https://mail.google.com/mail/", "Url for gmail (useful if you are logged into multiple accounts).")
}

var webCmd = &cobra.Command{
  Use:   "web",
  Short: "Start web server",
  Long:  `Starts up the Calliope web app.`,
  Run: func(cmd *cobra.Command, args []string) {
    web()
  },
}

type Data struct {
  Title string
  Report template.HTML
  Earliest time.Time
  Latest time.Time
  TotalEmails int64
}

func web() {
  options := report.Options{
    Label: label,
    Starred: !allMessages,
    InboxUrl: inboxUrl,
    Size: size,
  }

  client := misc.GetStoreClient()

  http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
    ShowReport(client, options, r, w)
  })

  fmt.Printf("Starting web server: http://localhost:%s/\n\n", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}

func ShowReport(client *store.Service, options report.Options, r *http.Request, w http.ResponseWriter) {
  var reportBuffer bytes.Buffer
  report.Run(client, &reportBuffer, options)

  t := template.Must(template.ParseFiles(
    "templates/layout.html",
    "templates/web-ui.html",
    ))
  queryJson := r.FormValue("query")
  fmt.Println(queryJson)
  stats, _ := client.GetStats()
  reportHtml := template.HTML(reportBuffer.String())
  data := Data{
    Title:       "Calliope Email Report",
    Report:      reportHtml,
    Earliest:    stats.Earliest,
    Latest:      stats.Latest,
    TotalEmails: stats.Total,
  }
  if err := t.ExecuteTemplate(w, "layout", data); err != nil {
    log.Println("Error occurred while executing template: ", err)
  }
}
