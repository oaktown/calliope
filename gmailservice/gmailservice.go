package gmailservice

import (
  "context"
  "fmt"
  // "bytes"
  "log"
  "encoding/json"
  "net/http"
  "google.golang.org/api/gmail/v1"
  "encoding/base64"
  // "github.com/kr/pretty"
)


// GmailService keeps state we need
type GmailService struct {
  svc         *gmail.Service
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

// Download doesn't do anything yet
func Download(g *GmailService, messages chan<- []byte) {
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
      msg, err := g.svc.Users.Messages.Get("me", m.Id).Do()
      if err != nil {
        log.Printf("Unable to retrieve message %v: %v", m.Id, err)
        continue
      }
      fmt.Printf("Sending Message ID: %v\n", m.Id)
      byt, _ := json.MarshalIndent(msg, "", "\t")
      messages <- byt
    }
    close(messages)
    return;
}

type GmailDoc struct {
  source []byte
}

type GmailMessagePart struct {
  Body struct {
    Data string `json:"data"`
  } `json:"body"`
  MimeType string `json:"mimeType"`
}

type GmailMessagePayload struct {
  Parts []GmailMessagePart `json:"parts"`
}

type GmailMessage struct {
  HistoryId string `json:"historyId"`
  Id string `json:"id"`
  InternalDate string `json:"internalDate"`
  LabelIds []string `json:labelIds`
  Payload GmailMessagePayload
}

func (doc *GmailDoc) JsonData() (GmailMessage, error) {
  var data GmailMessage
  if err := json.Unmarshal(doc.source, &data); err != nil {    
    log.Printf("json.Unmarshal failed, skipping message, err: %v", err)
    return data, err
  }
  return data, nil
}

func (doc *GmailDoc) BodyText() (string, error) {
  data, err := doc.JsonData()
  if err != nil {
    return "", err
  }
  parts := data.Payload.Parts
  for _, part := range parts {
    if part.MimeType == "text/plain" {
      encodedBody := part.Body.Data
      log.Printf("body: %v", encodedBody)
      body, _ := base64.URLEncoding.DecodeString(encodedBody)
      return string(body), nil
    }
  }
  return "", nil // TODO: is this the right thing to do when not found? Possibly should look at body field?
}


