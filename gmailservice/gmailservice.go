package gmailservice

import (
  "encoding/base64"
  "fmt"
  "log"
  "time"

  "google.golang.org/api/gmail/v1"
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

func Download(gmailService *gmail.Service, lastDate string, limit int64, pageToken string, inboxUrl string) ([]Message, error) {
  var messages []Message

  messagesToDownload, err := SearchMessages(gmailService, lastDate, limit, pageToken)
  if err != nil {
    log.Printf("Unable to retrieve messages: %v", err)
    return messages, err
  }

  messages = DownloadFullMessages(messagesToDownload, gmailService, inboxUrl)
  return messages, nil
}

func SearchMessages(svc *gmail.Service, after string, limit int64, pageToken string) ([]*gmail.Message, error) {
  request := svc.Users.Messages.List("me")
  if limit > 0 {
    request = request.MaxResults(limit)
  }
  if after != "" {
    request = request.Q("after: " + after)
  }
  if pageToken != "" {
    request = request.PageToken(pageToken)
  }
  response, err := request.Do()

  return response.Messages, err
}

func DownloadFullMessages(gmailMessages []*gmail.Message, svc *gmail.Service, inboxUrl string) []Message {
  var fullMessages []Message
  for _, m := range gmailMessages {
    message, err := DownloadFullMessage(svc, m.Id, inboxUrl)
    if err != nil {
      log.Printf("Could not download message with id %v. Error:\n\n%v\n\n", id, err)
    } else {
      fullMessages = append(fullMessages, message)
    }
  }
  return fullMessages
}

func DownloadFullMessage(svc *gmail.Service, id string, inboxUrl string) (Message, error) {
  gmailMsg, err := svc.Users.Messages.Get("me", id).Do()
  if err != nil {
    log.Printf("Unable to retrieve message %v: %v", id, err)
    return Message{}, err
  }
  fmt.Printf("Sending Message ID: %v\n", id)
  message, err := GmailToMessage(*gmailMsg, inboxUrl)
  return message, nil
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
    Id:      gmail.Id,
    Url:     fmt.Sprint(inboxUrl, gmail.ThreadId),
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
