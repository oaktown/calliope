package cmd

import (
  "fmt"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/web"
  "github.com/spf13/cobra"
  "log"
  "net/http"
  "strconv"
)

var port string
var allMessages bool

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
    label := r.FormValue("label")
    fmt.Println("starred:", r.FormValue("starred"))
    starred := r.FormValue("starred") == "true"
    inboxUrl := r.FormValue("gmailurl")
    sortField := r.FormValue("sort")
    sortAscending := r.FormValue("ascending") == "true"
    if inboxUrl == "" {
      inboxUrl = "https://mail.google.com/mail/"
    }
    size, err := strconv.Atoi(r.FormValue("size"))
    if err != nil {
      size = 100
    }

    opt := web.Options{
      Label:         label,
      Starred:       starred,
      InboxUrl:      inboxUrl,
      Size:          size,
      SortField:     sortField,
      SortAscending: sortAscending,
    }

    log.Printf("options: %+v\n", opt)
    var query string
    if query = r.FormValue("query"); query == "" {
      query = web.QueryStringFromLabel(client, opt)
    } else {
      query = r.FormValue("query")
    }

    log.Println("Query (web.go):", query)
    web.ShowHomePage(client, query, w, opt)
  })

  fmt.Printf("Starting web server: http://localhost:%s/\n\n", port)
  log.Fatal(http.ListenAndServe(":"+port, nil))
}
