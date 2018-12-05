package report

import (
  "fmt"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
  "html/template"
  "log"
  "os"
)

type Options struct {
  Label    string
  Starred  bool
  Query    string
  InboxUrl string
}

type Data struct {
  Label    string
  Messages []*gmailservice.Message
}

func Run(s *store.Service, options Options) {

  gmailUrl := func(threadId string) string {
    return fmt.Sprintf("%v#inbox/%v", options.InboxUrl, threadId)
  }

  jump := func(id string) string {
    return fmt.Sprint("#", id)
  }

  messages, err := s.GetMessages(options.Label, options.Starred)
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
      ParseFiles("report.html"))
  data := Data{
    Label:    options.Label,
    Messages: messages,
  }
  if err := report.Execute(os.Stdout, data); err != nil {
    log.Println("Error rendering template: ", err)
  }
}

