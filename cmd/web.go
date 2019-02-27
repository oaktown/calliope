package cmd

import (
  "encoding/json"
  "fmt"
  "github.com/gorilla/mux"
  "time"

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

  r := mux.NewRouter()

  r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("public"))))

  r.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
    web.ShowStats(r, w)
  })

  r.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
    svc := misc.GetStoreClient()
    options := getOptionsFromRequest(r)
    reportData := report.GetJsonReport(options, svc)
    // TODO: handle error
    reportJson, _ := json.MarshalIndent(reportData, "", "  ")
    w.Header().Set("Content-Type", "application/json")

    fmt.Fprint(w, string(reportJson))
  })

  r.HandleFunc("/message/{id:[^/]+}", func(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    web.ShowMessage(id, w, r)
  })

  r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
      fmt.Fprint(w, "API server started. Ready to accept connections.")
    } else {
      w.WriteHeader(http.StatusNotFound)
    }
  })

  srv := &http.Server{
    Handler:      r,
    Addr:         "127.0.0.1:8080",
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }

  fmt.Printf("Starting web server: http://localhost:%s/\n\n", port)
  log.Fatal(srv.ListenAndServe())
}

func getOptionsFromRequest(r *http.Request) report.QueryOptions {
  var sortField string
  if sortField = r.FormValue("sort"); sortField == "" {
    sortField = "Date"
  }
  inboxUrl := r.FormValue("gmailUrl")
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
