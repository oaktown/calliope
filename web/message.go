package web

import (
  "encoding/base64"
  "github.com/gorilla/mux"
  "github.com/microcosm-cc/bluemonday"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/store"
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

func showMessage(id string, w http.ResponseWriter, r *http.Request) {
  store := misc.GetStoreClient()
  message, err := store.GetMessage(id)
  if err != nil {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  data := MessageHtml{"Message id:" + id, getMessageBody(message)}

  t := template.Must(template.ParseFiles("templates/layout.html", "templates/message.html"))
  if err := t.ExecuteTemplate(w, "layout", data); err != nil {
    log.Println("Error occurred while executing status template: ", err)
  }
}

func getMessageBody(message store.Message) template.HTML {
  htmlBodyEncoded := gmailservice.GetBodyPartByMimeType(message.Source, "text/html")
  htmlBody, _ := base64.URLEncoding.DecodeString(htmlBodyEncoded)
  var unsafeHtml string
  if len(htmlBody) > 0 {
    unsafeHtml = string(htmlBody)
  } else {
    unsafeHtml ="<pre>\n" + message.Body + "\n</pre>"
  }

  p := bluemonday.UGCPolicy()
  html := p.Sanitize(unsafeHtml)
  return template.HTML(html)
}
