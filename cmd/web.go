package cmd

import (
  "encoding/json"
  "fmt"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
  "github.com/oaktown/calliope/web"
  "github.com/spf13/cobra"
  "log"
  "net/http"
  "strconv"
)

var port string

func init() {
  rootCmd.AddCommand(webCmd)
  webCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run web server on.")
}

var webCmd = &cobra.Command{
  Use:   "web",
  Short: "Start web server",
  Long:  `Starts up the Calliope web app.`,
  Run: func(cmd *cobra.Command, args []string) {
    startServer()
  },
}

func startServer() {
  http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "images/email_monster_Uxt_icon.ico")
  })

  http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
    web.ShowStats(r, w)
  })

  fs := http.FileServer(http.Dir("public/assets"))
  http.Handle("/assets/", http.StripPrefix("/assets/", fs))

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
      options:= getOptionsFromRequest(r)
      web.ShowHomepage(r, w, options)
    } else {
      fmt.Fprint(w, "Nothing to see here.")
    }
  })

  http.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
    svc := misc.GetStoreClient()
    options:= getOptionsFromRequest(r)
    reportData:= report.GetJsonReport(options, svc)
    // TODO: handle error
    reportJson, _ := json.MarshalIndent(reportData, "", "  ")
    w.Header().Set("Content-Type", "application/json")

    fmt.Fprint(w, string(reportJson))
  })

  fmt.Printf("Starting web server: http://localhost:%s/\n\n", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getOptionsFromRequest(r *http.Request) report.QueryOptions {
  var sortField string
  if sortField = r.FormValue("sort"); sortField == "" {
    sortField = "Date"
  }
  inboxUrl := r.FormValue("gmailurl")
  if inboxUrl == "" {
    inboxUrl = "https://mail.google.com/mail/"
  }
  size, err := strconv.Atoi(r.FormValue("size"))
  if err != nil {
    size = 100
  }
  var timezone string
  if timezone = r.FormValue("timezone"); timezone == "" {
    timezone = "-0800" // Default to PST
  }
  opt := report.QueryOptions{
    StartDate:     r.FormValue("startDate"),
    EndDate:       r.FormValue("endDate"),
    Timezone:      timezone,
    Participants:  r.FormValue("participants"),
    BodyOrSubject: r.FormValue("bodyOrSubject"),
    Label:         r.FormValue("label"),
    Starred:       r.FormValue("starred") == "true",
    InboxUrl:      inboxUrl,
    Size:          size,
    SortField:     sortField,
    SortAscending: r.FormValue("ascending") == "true",
    Query:         r.FormValue("query"),
  }
  return opt
}
