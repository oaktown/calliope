package cmd

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
  "github.com/oaktown/calliope/store"
  "github.com/olivere/elastic"
  "github.com/spf13/cobra"
  "html/template"
  "log"
  "net/http"
  "time"
)

var port, label, url string
var size int
var allMessages bool


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
  Title       string
  Size        int
  Query       string
  Report      template.HTML
  Earliest    time.Time
  Latest      time.Time
  TotalEmails int64
}

func web() {
  starred := !allMessages
  client := misc.GetStoreClient()
  initialQuery := true

  http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "images/email_monster_Uxt_icon.ico")
  })

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    var query string
    if initialQuery {
      query = GetInitialQueryString(client, starred)
      initialQuery = false
    } else {
      query = r.FormValue("query")
    }

    ShowHomePage(client, query, w)
  })

  fmt.Printf("Starting web server: http://localhost:%s/\n\n", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}

func ShowHomePage(client *store.Service, query string, w http.ResponseWriter) {
  stats, _ := client.GetStats()
  reportHtml := template.HTML(GetReportHtml(client, query))
  data := Data{
    Title:       "Calliope Email Report",
    Size:        size,
    Query:       query,
    Report:      reportHtml,
    Earliest:    stats.Earliest,
    Latest:      stats.Latest,
    TotalEmails: stats.Total,
  }

  t := template.Must(template.ParseFiles("templates/layout.html", "templates/web-ui.html"))

  if err := t.ExecuteTemplate(w, "layout", data); err != nil {
    log.Println("Error occurred while executing template: ", err)
  }
}

func GetReportHtml(client *store.Service, query string) string {
  var req *elastic.SearchService
  req = client.GetRawQuery(query, size)
  var reportBuffer bytes.Buffer
  report.Run(client, &reportBuffer, req, inboxUrl)
  return reportBuffer.String()
}

func GetInitialQueryString(client *store.Service, starred bool) string {
  boolQuery, _ := client.GetQueryFromLabel(label, starred, size)
  source, _ := boolQuery.Source()
  q, _ := json.MarshalIndent(source, "  ", "  ")
  query = fmt.Sprintf("{\n  \"query\": %s\n}", q)
  fmt.Println("Query using label: ", query)
  return query
}
