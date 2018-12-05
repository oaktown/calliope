package report

import (
  "html/template"
  "os"

  "google.golang.org/api/gmail/v1"
)

func Run(gmail *gmail.Service, label string) {
  report := template.Must(template.ParseFiles("report.html"))
  obj := struct{
    foo string
    bar int
  }{"baz", 1}
  report.Execute(os.Stdout, obj)
}

//URL:
//
//https://mail.google.com/mail/u/1/#search/{{query}}/{{threadid}}
// As opposed to just linking to the thread, if you click on the back arrow, it'll put you in the search
// for the label + star. Also, if there is a search term, it will highlight it.






