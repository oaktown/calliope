package report

import (
  "fmt"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
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
  Label    string
  Messages []*gmailservice.Message
}

func Run(s *store.Service, wr io.Writer, options Options) {

  gmailUrl := func(threadId string) string {
    return fmt.Sprintf("%v#inbox/%v", options.InboxUrl, threadId)
  }

  jump := func(id string) string {
    return fmt.Sprint("#", id)
  }

  messages, err := s.GetMessages(options.Label, options.Starred, options.Size)
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
    Label:    options.Label,
    Messages: messages,
  }
  if err := report.Execute(wr, data); err != nil {
    log.Println("Error rendering template: ", err)
  }
}
