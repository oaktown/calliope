package gmailservice

import (
  "context"
  "fmt"
  "log"
  "encoding/json"
  "net/http"
  "google.golang.org/api/gmail/v1"
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



func (doc *GmailDoc) JsonData() (map[string]interface{}) {
  // should memoize
  data := make(map[string]interface{})
  log.Printf("doc.source: %v", string(doc.source))
  if err := json.Unmarshal(doc.source, &data); err != nil {    
    log.Printf("json.Unmarshal failed, skipping message, err: %v", err)
    return nil;
  }
  return data
}

func (doc *GmailDoc) BodyText() string {
  data := doc.JsonData()
  parts := (data["payload"].(map[string]interface{}))["parts"]
  log.Printf("parts: %v", parts)
  for _, part := range parts.([]interface{}) {
    part := part.(map[string]interface{})
    if part["mimeType"] == "text/plain" {
      return part["body"].(map[string]interface{})["data"].(string)
    }
  }
  return ""
  // part["mimeType"] == "text/plain"
  // part["body"]["data"] // decode this
  // return "something else"
}


