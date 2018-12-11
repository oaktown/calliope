package report

import (
  "fmt"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
  "github.com/olivere/elastic"
  "html/template"
  "io"
  "log"
)

type Options struct {
  Label    string
  Starred  bool
  InboxUrl string
  Size     int
}

type Data struct {
  Messages []*gmailservice.Message
}
//report.Run(client, &reportBuffer, req, inboxUrl)
func Run(s *store.Service, wr io.Writer, req *elastic.SearchService, inboxUrl string) {

  gmailUrl := func(threadId string) string {
    return fmt.Sprintf("%v#inbox/%v", inboxUrl, threadId)
  }

  jump := func(id string) string {
    return fmt.Sprint("#", id)
  }

  messages, err := s.GetMessages(req)
  if err != nil {
    log.Println("Exiting due to error")
    return
  }

  report := template.Must(
    template.New("report.html").
      Funcs(template.FuncMap{
        "gmailUrl": gmailUrl,
        "jump":     jump,
      }).
      ParseFiles("templates/report.html"))
  data := Data{
    Messages: messages,
  }
  if err := report.Execute(wr, data); err != nil {
    log.Println("Error rendering template: ", err)
  }
}
