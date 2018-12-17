package web

import (
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
  "github.com/oaktown/calliope/store"
  "html/template"
  "log"
  "net/http"
  "strconv"
)

type HomepageData struct {
  Title       string
  Fields      report.QueryOptions
  Size        int
  Query       string
  Report      template.HTML
}

type Homepage struct {
  r   *http.Request
  w   http.ResponseWriter
  svc *store.Service
}

func ShowHomepage(r *http.Request, w http.ResponseWriter) {
  h := Homepage{
    r:   r,
    w:   w,
    svc: misc.GetStoreClient(),
  }
  formFields := h.getFormFields()
  h.render(formFields)
}

func (h Homepage) getFormFields() report.QueryOptions {
  var sortField string
  r := h.r
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

func (h Homepage) render(opt report.QueryOptions) {
  rpt := report.GetReport(opt, h.svc)

  data := HomepageData{
    Title:       "Calliope Email Report",
    Fields:      opt,
    Query:       rpt.Query,
    Report:      rpt.Html,
  }

  t := template.Must(template.ParseFiles("templates/layout.html", "templates/homepage.html"))

  if err := t.ExecuteTemplate(h.w, "layout", data); err != nil {
    log.Println("Error occurred while executing homepage template: ", err)
  }
}
