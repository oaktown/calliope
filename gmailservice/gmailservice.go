package gmailservice

import (
  "context"
  "fmt"
	"time"
  "log"
  "encoding/json"
  "net/http"
  "google.golang.org/api/gmail/v1"
	"encoding/base64"
)

type Message struct {
	Id string
	Date time.Time
	To string
	Cc string
	From string
	Subject string
	Body string // the thing we're decoding
	Source gmail.Message
}

// TODO: Remove this type
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

// TODO: 1. Feature: Channel is sent Messages (e.g. w/ decoded body)
// TODO: 2. Refactor: Remove the use of a channel – this function should return []Message
// TODO (cont'd) and lastDate, pageToken, and batchSize should be parameters
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

func JsonToGmail(jsonByteArray []byte) (gmail.Message, error) {
  var data gmail.Message
  if err := json.Unmarshal(jsonByteArray, &data); err != nil {
    log.Printf("json.Unmarshal failed, skipping message, err: %v", err)
    return data, err
  }
  return data, nil
}

// TODO: BodyText could be in the body field instead of in the payload. We
// should probably add logic to handle that case, or at least log or something. 
func BodyText(msg gmail.Message) (string) {
  parts := msg.Payload.Parts
  for _, part := range parts {
    if part.MimeType == "text/plain" {
      encodedBody := part.Body.Data
      body, _ := base64.URLEncoding.DecodeString(encodedBody)
      return string(body)
    }
	}
  return ""
}

func GmailToMessage(gmail gmail.Message) (Message, error) {
	date := time.Unix(gmail.InternalDate / 1000, 0)
	body := BodyText(gmail)
	message := Message {
		Id: gmail.Id,
		Date: date,
		To: "",
		Cc: "",
		From: "",
		Subject: "",
  	Body: body,
		Source: gmail,
	}	
	return message, nil
}
