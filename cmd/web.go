package cmd

import (
  "fmt"
  "github.com/gorilla/mux"
  "time"

  "github.com/oaktown/calliope/api"
  "github.com/oaktown/calliope/web"
  "github.com/spf13/cobra"
  "log"
  "net/http"
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
  r.HandleFunc("/stats", web.StatsHandler)
  r.HandleFunc("/api/search", api.SearchHandler)
  r.HandleFunc("/message/{id:[^/]+}", web.MessageHandler)
  r.HandleFunc("/report", web.ReportHandler)
  r.HandleFunc("/", DefaultHandler)

  srv := &http.Server{
    Handler:      r,
    Addr:         "127.0.0.1:8080",
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }

  fmt.Printf("Starting web server: http://localhost:%s/\n\n", port)
  log.Fatal(srv.ListenAndServe())
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "/" {
    fmt.Fprint(w, "API server started. Ready to accept connections.")
  } else {
    w.WriteHeader(http.StatusNotFound)
  }
}

