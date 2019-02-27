package api

import (
  "encoding/json"
  "fmt"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/report"
  "net/http"
  "strconv"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
  svc := misc.GetStoreClient()
  options := searchOptionsFromParams(r)
  reportData := report.GetJsonReport(options, svc)
  // TODO: handle error
  reportJson, _ := json.MarshalIndent(reportData, "", "  ")
  w.Header().Set("Content-Type", "application/json")

  fmt.Fprint(w, string(reportJson))
}

func searchOptionsFromParams(r *http.Request) report.QueryOptions {
  var sortField string
  if sortField = r.FormValue("sort"); sortField == "" {
    sortField = "Date"
  }
  inboxUrl := r.FormValue("gmailUrl")
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
