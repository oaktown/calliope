package gmailservice

import (
  "encoding/base64"
  "fmt"
  "log"
  "time"

  "google.golang.org/api/gmail/v1"
)

type Message struct {
  Id       string
  Url      string
  Date     time.Time
  To       string
  Cc       string
  From     string
  Subject  string
  Body     string
  ThreadId string
  Snippet  string
  Source   gmail.Message
}

type Downloader struct {
  SearchChan   chan *gmail.Message
  MessageChan  chan *Message
  M2           chan *Message
  WorkersQueue chan bool
  MaxWorkers   int
  Svc          *gmail.Service
  Options      Options
  DoList       func(*gmail.UsersMessagesListCall) (*gmail.ListMessagesResponse, error)
  DoGet        func(request *gmail.UsersMessagesGetCall) (*gmail.Message, error)
}

type Options struct {
  Query    string
  Limit    int64
  InboxUrl string
}

func New(svc *gmail.Service, options Options, maxWorkers int) Downloader {
  search := make(chan *gmail.Message)
  message := make(chan *Message)
  workers := make(chan bool, maxWorkers)
  return Downloader{
    SearchChan:   search,
    MessageChan:  message,
    WorkersQueue: workers,
    MaxWorkers:   maxWorkers,
    Svc:          svc,
    Options:      options,
    DoList:       DoList,
    DoGet:        DoGet,
  }
}

func (d Downloader) NoNewWorkers() {
  for i := 0; i < d.MaxWorkers; i++ {
    d.WorkersQueue <- true
  }
}

// Download everything that is requested in calliope generic Message format
func Download(d Downloader) {
  go SearchMessages(d)
  go DownloadFullMessages(d)
}

// SearchMessages gets list of message and thread IDs (not full message content)
func SearchMessages(d Downloader) {
  var totalMessages int64
  totalMessages = 0
  // Seems like MaxResults over 500 results in pages of 500; possibly subject to change?
  request := d.Svc.Users.Messages.List("me")
  if d.Options.Limit > 0 {
    request = request.MaxResults(d.Options.Limit)
  }
  if d.Options.Query != "" {
    log.Println("adding: ", d.Options.Query)
    request = request.Q(d.Options.Query)
  }
  pageToken := ""
  for {
    if pageToken != "" {
      request = request.PageToken(pageToken)
    }
    response, err := d.DoList(request)

    if err != nil {
      // Could be transient error (e.g. throttle), but for now we're just exiting
      log.Print("search error: ", err)
      return
    }
    pageToken = response.NextPageToken
    for _, result := range response.Messages {
      d.SearchChan <- result
    }
    totalMessages += int64(len(response.Messages))
    log.Printf("NextPageToken: %v\nEstimate: %v\nMessages found this page: %v\n\n", pageToken, response.ResultSizeEstimate, len(response.Messages))
    if pageToken == "" || totalMessages >= d.Options.Limit {
      break
    }
  }
  close(d.SearchChan)
}

func DoList(request *gmail.UsersMessagesListCall) (*gmail.ListMessagesResponse, error) {
  response, err := request.Do()
  return response, err
}

func DownloadFullMessages(d Downloader) {
  defer close(d.MessageChan)
  for searchResult := range d.SearchChan {
    d.WorkersQueue <- true
    go DownloadFullMessage(d, searchResult.Id)
  }
  d.NoNewWorkers()
}

func DownloadFullMessage(d Downloader, id string) {
  defer func() { <-d.WorkersQueue }()
  request := d.Svc.Users.Messages.Get("me", id)
  gmailMsg, err := d.DoGet(request)
  log.Println("Fetching message id:", id)
  if err != nil {
    log.Printf("Unable to retrieve message %v: %v", id, err)
    return
  }
  message, _ := GmailToMessage(*gmailMsg, d.Options.InboxUrl)
  if err != nil {
    log.Printf("Unable to decode message %v: %v", id, err)
    return
  }
  log.Println("Subject: ", message.Subject)
  d.MessageChan <- &message
}

func DoGet(request *gmail.UsersMessagesGetCall) (*gmail.Message, error) {
  gmailMsg, err := request.Do()
  return gmailMsg, err
}

func BodyText(msg gmail.Message) string {
  // TODO: We might want to see if there are other places the body can be located.
  parts := msg.Payload.Parts
  if msg.Payload.Body.Data != "" {
    body, _ := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
    return string(body)
  } else {
    for _, part := range parts {
      if part.MimeType == "text/plain" {
        encodedBody := part.Body.Data
        body, _ := base64.URLEncoding.DecodeString(encodedBody)
        return string(body)
      }
    }
  }
  return ""
}

func ExtractHeader(gmail gmail.Message, field string) string {
  // TODO: For now, we just grab the first one, but really we should probably
  // figure out which one is the significant one, or if they should be merged, etc.
  // This probably doesn't apply to all headers, but some might repeat (or maybe
  // that's only in original, and Gmail makes some decision about which one should
  // win)
  for _, header := range gmail.Payload.Headers {
    if header.Name == field {
      return header.Value
    }
  }
  return ""
}

func GmailToMessage(gmail gmail.Message, inboxUrl string) (Message, error) {
  // TODO: decode all of the fields, not just plain-text body
  date := time.Unix(gmail.InternalDate/1000, 0)
  body := BodyText(gmail)
  message := Message{
    Id:       gmail.Id,
    Url:      fmt.Sprintf("%v#inbox/%v", inboxUrl, gmail.ThreadId),
    Date:     date,
    To:       ExtractHeader(gmail, "To"),
    Cc:       ExtractHeader(gmail, "Cc"),
    From:     ExtractHeader(gmail, "From"),
    Subject:  ExtractHeader(gmail, "Subject"),
    Body:     body,
    ThreadId: gmail.ThreadId,
    Snippet:  gmail.Snippet,
    Source:   gmail,
  }
  return message, nil
}
