package gmailservice

import (
  "context"
  "encoding/base64"
  "encoding/json"
  "fmt"
  "google.golang.org/api/gmail/v1"
  "log"
  "net/http"
  "os"
  "time"
)

type Message struct {
  Id      string
  Url     string
  Date    time.Time
  To      string
  Cc      string
  From    string
  Subject string
  Body    string // the thing we're decoding
  Source  gmail.Message
}

// TODO: Remove this type
type GmailService struct {
  svc *gmail.Service
}

// New returns GmailService initialized with given client
func New(ctx context.Context, client *http.Client) (*GmailService, error) {
  g := new(GmailService)
  svc, err := gmail.New(client)
  if err != nil {
    log.Printf("could not create gmail client, %v", err)
    return nil, err
  }
  g.svc = svc;
  return g, nil
}

// TODO: Refactor: Remove the use of a channel – this function should return
// []Message. Also, lastDate, pageToken, and batchSize should be parameters
func Download(g *GmailService, messages chan<- Message) {
  lastDate := "2018/01/01"
  var pageToken = ""

  var req *gmail.UsersMessagesListCall

  if lastDate == "" {
    log.Println("Retrieving all messages.")
    req = g.svc.Users.Messages.List("me")

  } else {
    log.Println("Retrieving messages starting on", lastDate)
    req = g.svc.Users.Messages.List("me").Q("after: " + lastDate)
  }

  if pageToken != "" {
    req.PageToken(pageToken)
  }
  r, err := req.Do()

  if err != nil {
    log.Printf("Unable to retrieve messages: %v", err)
    return
    //continue
  }

  log.Printf("Processing %v messages...\n", len(r.Messages))

  for _, m := range r.Messages[:6] {
    gmailMsg, err := g.svc.Users.Messages.Get("me", m.Id).Do()
    if err != nil {
      log.Printf("Unable to retrieve message %v: %v", m.Id, err)
      continue
    }
    fmt.Printf("Sending Message ID: %v\n", m.Id)
    message, err := GmailToMessage(*gmailMsg)
    // byt, _ := json.MarshalIndent(message, "", "\t")
    messages <- message
  }
  close(messages)
  return
}

func JsonToGmail(jsonByteArray []byte) (gmail.Message, error) {
  var data gmail.Message
  if err := json.Unmarshal(jsonByteArray, &data); err != nil {
    log.Printf("json.Unmarshal failed, skipping message, err: %v", err)
    return data, err
  }
  return data, nil
}

// TODO: We might want to see if there are other places the body can be located.
func BodyText(msg gmail.Message) (string) {
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

func ExtractHeader(gmail gmail.Message, field string) (string) {
  // TODO: For now, we just grab the first one, but really we should probably
  // figure out which one is the signficant one, or if they should be merged, etc.
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

func GmailToMessage(gmail gmail.Message) (Message, error) {
  date := time.Unix(gmail.InternalDate/1000, 0)
  body := BodyText(gmail)
  // TODO: If we do this at all, we should probably have the value passed in
  // instead of reading from env var in this package
  gmailUserStr := os.Getenv("GMAIL_USER_STRING")
  message := Message{
    Id:      gmail.Id,
    Url:     fmt.Sprintf("https://mail.google.com/mail/%v#inbox/%v", gmailUserStr, gmail.ThreadId),
    Date:    date,
    To:      ExtractHeader(gmail, "To"),
    Cc:      ExtractHeader(gmail, "Cc"),
    From:    ExtractHeader(gmail, "From"),
    Subject: ExtractHeader(gmail, "Subject"),
    Body:    body,
    Source:  gmail,
  }
  return message, nil
}
