package cmd

import (
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

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    log.Println("Request path:", r.URL.Path)
    client := misc.GetStoreClient()
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

    opt := report.QueryOptions{
      StartDate:     r.FormValue("startDate"),
      EndDate:       r.FormValue("endDate"),
      Participants:  r.FormValue("participants"),
      Label:         r.FormValue("label"),
      Starred:       r.FormValue("starred") == "true",
      InboxUrl:      inboxUrl,
      Size:          size,
      SortField:     sortField,
      SortAscending: r.FormValue("ascending") == "true",
      Query: r.FormValue("query"),
    }
    log.Printf("options: %+v\n", opt)

    web.ShowHomePage(client, w, opt)
  })

  fmt.Printf("Starting web server: http://localhost:%s/\n\n", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}
