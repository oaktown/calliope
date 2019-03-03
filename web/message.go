package web

import (
  "github.com/gorilla/mux"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
  "html/template"
  "log"
  "net/http"
)

type MessageHtml struct{
  Title string
  Body template.HTML
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
  id := mux.Vars(r)["id"]
  showMessage(id, w, r)
}

func showMessage(id string, w http.ResponseWriter, _ *http.Request) {
  store := misc.GetStoreClient()
  message, err := store.GetMessage(id)
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  data := MessageHtml{
    Title:"Message id:" + id,
    Body:template.HTML(report.GetMessageHtmlBody(message)),
  }

  t := template.Must(template.ParseFiles("templates/layout.html", "templates/message.html"))
  if err := t.ExecuteTemplate(w, "layout", data); err != nil {
    log.Println("Error occurred while executing status template: ", err)
  }
}
