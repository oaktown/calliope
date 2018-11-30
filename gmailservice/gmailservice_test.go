package gmailservice

import (
  "encoding/json"
  "google.golang.org/api/gmail/v1"
  "testing"
  "io/ioutil"
  // "encoding/json"
  "log"
  "os"
  "strings"
)

var expectedBody = "How are you doing? Some special characters: <>?$/:-)"

func getEmailJson() []byte {
  jsonFile, err := os.Open("gmail_message.json") // Slack notification

  if err != nil {
    log.Fatalf("Couldn't open gmail_message.json. Error: %v", err)
  }

  emailJson, _ := ioutil.ReadAll(jsonFile) // theoretically the same thing that the gmail API returns
  jsonFile.Close()
  return emailJson
}

func JsonToGmail(jsonByteArray []byte) (gmail.Message, error) {
  var data gmail.Message
  if err := json.Unmarshal(jsonByteArray, &data); err != nil {
    log.Printf("json.Unmarshal failed, skipping message, err: %v", err)
    return data, err
  }
  return data, nil
}

func TestDownload(t *testing.T) {
  t.Skip()
}

func TestGetIndexOfMessages(t *testing.T) {
  t.Skip()
}

func TestDownloadFullMessages(t *testing.T) {
  t.Skip()
}

func TestExtractHeader(t *testing.T) {
  t.Skip()
}

func TestBodyText(t *testing.T) {
  emailJson := getEmailJson()
  gmailMsg, _ := JsonToGmail(emailJson)

  if body := BodyText(gmailMsg); !strings.Contains(body, expectedBody) {
    t.Errorf("Body is incorrect. Should have been:\n\n%v\n\nInstead, got:\n\n%v\n\n", expectedBody, body)
  }
}

func TestGmailToMessage(t *testing.T) {
  emailJson := getEmailJson()
  rawGmail, _ := JsonToGmail(emailJson)
  msg, err := GmailToMessage(rawGmail, "https://mail.google.com/mail/#inbox")
  if err != nil {
    t.Errorf("Unexpected error calling JsonForElasticsearch: %v", err)
  }
  if body := msg.Body; !strings.Contains(body, expectedBody) {
    t.Errorf("Body is incorrect. Should have been:\n\n%v\n\nInstead, got:\n\n%v\n\n", expectedBody, body)
  }
}
