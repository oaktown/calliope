package gmailservice

import (
	"encoding/json"
	"google.golang.org/api/gmail/v1"
	"testing"
	// "encoding/json"
	"log"
	"io/ioutil"
	"os"
	"strings"
)

func getEmailJson() []byte {
	jsonFile, err := os.Open("../gmail_message.json") // Slack notification

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

func TestJsonToGmail (t *testing.T) {
	emailJson := getEmailJson()
	gmailMsg, _ := JsonToGmail(emailJson)
	expected := "16753a2ae7f95997"
	if gmailMsg.Id != expected {
		t.Errorf("Expected %v but got: %v", expected, gmailMsg.Id)
	}
}

func TestBodyText(t *testing.T) {
	emailJson := getEmailJson()
	gmailMsg, _ := JsonToGmail(emailJson)

	if body := BodyText(gmailMsg); !strings.Contains(body, "ultrasaurus") {
		t.Errorf("Body was not correctly decoded. Should have an ultrasaurus. Instead got:\n\n%v\n\n", body)
	}
}

func TestGmailToMessage(t *testing.T) {
	emailJson := getEmailJson()
	rawGmail, _ := JsonToGmail(emailJson)
	msg, err := GmailToMessage(rawGmail, "https://mail.google.com/mail/#inbox")
	if err != nil {
		t.Errorf("Unexpected error calling JsonForElasticsearch: %v", err)
	}
	if body := msg.Body; !strings.Contains(body, "ultrasaurus") {
		t.Errorf("Body is incorrect. Should have an ultrasaurus. Instead got:\n\n%v\n\n", body)
	}
}
