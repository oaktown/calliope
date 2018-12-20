package gmailservice

import (
  "encoding/base64"
  "fmt"
  "github.com/jonboulle/clockwork"
  "github.com/oaktown/calliope/store"
  "google.golang.org/api/gmail/v1"
  "google.golang.org/api/googleapi"
  "log"
  "strings"
  "time"
)

type Downloader struct {
  SearchChan     chan *gmail.Message
  MessageChan    chan *store.Message
  M2             chan *store.Message
  WorkersQueue   chan bool
  MaxWorkers     int
  Svc            *gmail.Service
  Options        Options
  doList         func(*gmail.UsersMessagesListCall) (*gmail.ListMessagesResponse, error)
  doGet          func(*Downloader, string) (*gmail.Message, error)
  DoListLabels   func(*gmail.UsersLabelsListCall) ()
  GmailToMessage func(gmail.Message, string, time.Time) (store.Message, error)
  StartedAt      time.Time
  clock          clockwork.Clock
}

type Options struct {
  Query          string
  Limit          int64
  InboxUrl       string
  ExcludeHeaders map[string][]string
}

func New(svc *gmail.Service, options Options, maxWorkers int) Downloader {
  search := make(chan *gmail.Message)
  message := make(chan *store.Message)
  workers := make(chan bool, maxWorkers)
  return Downloader{
    SearchChan:     search,
    MessageChan:    message,
    WorkersQueue:   workers,
    MaxWorkers:     maxWorkers,
    Svc:            svc,
    Options:        options,
    doList:         doList,
    doGet:          doGet,
    GmailToMessage: GmailToMessage,
    StartedAt:      time.Now(),
  }
}

func (d Downloader) NoNewWorkers() {
  for i := 0; i < d.MaxWorkers; i++ {
    d.WorkersQueue <- true
  }
}

// Download everything that is requested in calliope generic Message format
func Download(d Downloader) []*store.Label {
  labels := DownloadLabels(d)
  go SearchMessages(d)
  go DownloadFullMessages(d)
  return labels
}

func DownloadLabels(d Downloader) []*store.Label {
  request := d.Svc.Users.Labels.List("me")
  response, _ := request.Do()
  var labels []*store.Label
  for _, l := range response.Labels {
    label := &store.Label{
      Id:   l.Id,
      Name: l.Name,
    }
    labels = append(labels, label)
  }
  return labels
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
    request = request.Q(d.Options.Query)
  }
  pageToken := ""
  for {
    if pageToken != "" {
      request = request.PageToken(pageToken)
    }
    response, err := d.doListWrapper(request)

    if err != nil {
      log.Printf("!!!!!!!!!!!!!!!!!!!!! Error calling search API. \n  Query: %v \n  Page token: %v\n  Search error: %v", d.Options.Query, pageToken, err)
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

func (d *Downloader) doListWrapper(request *gmail.UsersMessagesListCall) (*gmail.ListMessagesResponse, error) {
  var response *gmail.ListMessagesResponse
  fn := func() error {
    r, err := doList(request)
    response = r
    return err
  }
  err := d.tryThrice(fn)
  return response, err
}

func doList(request *gmail.UsersMessagesListCall) (*gmail.ListMessagesResponse, error) {
  return request.Do()
}

func DownloadFullMessages(d Downloader) {
  defer close(d.MessageChan)
  for searchResult := range d.SearchChan {
    d.WorkersQueue <- true
    go DownloadFullMessage(d, searchResult.Id)
  }
  d.NoNewWorkers()
}

func HasMatchingHeader(excludeHeaders map[string][]string, message gmail.Message) (string, string) {
  // If header matches, returns first matching header; otherwise returns empty strings
  for header, excludeValues := range excludeHeaders {
    value := strings.ToLower(ExtractHeader(message, header))
    if value == "" {
      continue // No matching header
    }
    for _, excludeValue := range excludeValues {
      excludeValue = strings.ToLower(excludeValue)
      if strings.Contains(value, excludeValue) {
        return header, value
      }
    }
  }

  return "", ""
}

func DownloadFullMessage(d Downloader, id string) {
  defer func() { <-d.WorkersQueue }()

  partialMessageWithError := func(errMsg string) {
    d.MessageChan <- &store.Message{
      Id:                  id,
      DownloadedStartedAt: d.StartedAt,
      Subject:             errMsg,
    }
    log.Print(errMsg)
  }

  //TODO: Can't call this because it triggers oauth. need to be able to stub it out, too
  gmailMsg, err := d.DoGetWrapper(id)
  log.Println("Fetching message id:", id)
  if err != nil {
    errMsg := fmt.Sprintf("Unable to retrieve message %v: %v", id, err)
    partialMessageWithError(errMsg)
    return
  }
  message, _ := d.GmailToMessage(*gmailMsg, d.Options.InboxUrl, d.StartedAt)
  if err != nil {
    errMsg := fmt.Sprintf("Unable to decode message %v: %v", id, err)
    partialMessageWithError(errMsg)
    return
  }
  header, value := HasMatchingHeader(d.Options.ExcludeHeaders, *gmailMsg)
  if header == "" {
    log.Printf("Downloaded message %v\n  Subject: %v\n", id, message.Subject)
    d.MessageChan <- &message
  } else {
    log.Printf("Skipping message: \n  Subject: %s\n  Matching header: %v: %v\n", message.Subject, header, value)
  }
}

func (d *Downloader) DoGetWrapper(id string) (*gmail.Message, error) {
  // This method exists mostly so tryThrice doesn't have to be implemented twice:
  // once for list and once for message (since they have different signatures).
  // Could have been part of DownloadFullMessage but that's long enough already.
  // Almost the same as DoListWrapper except for the type in closure.
  var gmailMsg *gmail.Message
  fn := func() error {
    r, err := d.doGet(d, id)
    gmailMsg = r
    return err
  }
  err := d.tryThrice(fn)
  return gmailMsg, err
}

func doGet(d *Downloader, id string) (*gmail.Message, error) {
  return d.Svc.Users.Messages.Get("me", id).Do()
}

func BodyText(msg gmail.Message) string {
  // TODO: We might want to see if there are other places the body can be located.
  parts := msg.Payload.Parts
  //TODO: when there is no body
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
    if strings.ToLower(header.Name) == strings.ToLower(field) {
      return header.Value
    }
  }
  return ""
}

func GmailToMessage(gmail gmail.Message, inboxUrl string, downloaded time.Time) (store.Message, error) {
  // TODO: decode all of the fields, not just plain-text body
  date := time.Unix(gmail.InternalDate/1000, 0)
  body := BodyText(gmail)
  message := store.Message{
    Id:                  gmail.Id,
    Url:                 fmt.Sprintf("%v#inbox/%v", inboxUrl, gmail.ThreadId),
    Date:                date,
    DownloadedStartedAt: downloaded,
    To:                  ExtractHeader(gmail, "To"),
    Cc:                  ExtractHeader(gmail, "Cc"),
    From:                ExtractHeader(gmail, "From"),
    Subject:             ExtractHeader(gmail, "Subject"),
    Body:                body,
    ThreadId:            gmail.ThreadId,
    LabelIds:            gmail.LabelIds,
    Snippet:             gmail.Snippet,
    Source:              gmail,
  }
  return message, nil
}

func (d *Downloader) tryThrice(fn func() error) error {
  var err error
  for count := 0; count < 3; count ++ {
    err = fn()
    if err == nil {
      break
    }
    code := err.(*googleapi.Error).Code
    // If we have exceeded API usage, API will return 429 or 403.
    if code == 429 || code == 403 {
      secs := 10.0 * (count + 1)
      duration, _ := time.ParseDuration(fmt.Sprintf("%ds", secs))
      d.clock.Sleep(duration)
    } else {
      // If it's something else, we don't know how to deal with it, so just pass on the error
      return err
    }
  }
  return err
}
