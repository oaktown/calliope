package web

import (
	"fmt"
	"github.com/oaktown/calliope/report"
	"github.com/oaktown/calliope/store"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/oaktown/calliope/misc"
)

type HtmlReport struct {
	Html  template.HTML
	Query string
}

type HtmlData struct {
	Title    string
	Messages []*report.MessageWithHtml
}

func ReportHandler(w http.ResponseWriter, r *http.Request) {
	svc := misc.GetStoreClient()
	inboxUrl := r.FormValue("gmailUrl")
	if inboxUrl == "" {
		inboxUrl = "https://mail.google.com/mail/"
	}
	size, err := strconv.Atoi(r.FormValue("size"))
	if err != nil {
		size = 100
	}
	label := r.FormValue("label")

	messageSearch := svc.NewStructuredMessageSearch().Label(label).Size(size)
	RenderReport(w, messageSearch, inboxUrl)
}

func RenderReport(wr io.Writer, messageSearch store.MessageSearch, inboxUrl string) {
	gmailUrl := func(threadId string) string {
		return fmt.Sprintf("%v#inbox/%v", inboxUrl, threadId)
	}

	jump := func(id string) string {
		return fmt.Sprint("#", id)
	}

	messages, err := messageSearch.Do()
	messagesWithHtml := report.FillInHtmlBody(messages)

	if err != nil {
		log.Fatalf("Exiting due to error: %+v", err)
		return
	}

	report := template.Must(
		template.New("report.html").
			Funcs(template.FuncMap{
				"gmailUrl": gmailUrl,
				"jump":     jump,
			}).
			ParseFiles("templates/layout.html", "templates/report.html"))
	data := HtmlData{
		Title:    "Report",
		Messages: messagesWithHtml,
	}
	log.Println("number of message: ", len(messages))

	if err := report.ExecuteTemplate(wr, "layout", data); err != nil {
		log.Println("Error rendering template: ", err)
	}
}
