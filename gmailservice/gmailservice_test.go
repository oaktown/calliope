package gmailservice

import (
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
	
	json, _ := ioutil.ReadAll(jsonFile) // theoretically the same thing that the gmail API returns
	// log.Printf("as a string: %v", string(json))
	jsonFile.Close()
	return json
}

func TestJsonToGmail (t *testing.T) {
	json := getEmailJson()
	gmail, _ := JsonToGmail(json)
	expected := "16753a2ae7f95997"
	if gmail.Id != expected {
		t.Errorf("Expected %v but got: %v", expected, gmail.Id)
	}
}

func TestBodyText(t *testing.T) {
	json := getEmailJson()
	gmail, _ := JsonToGmail(json)

	if body := BodyText(gmail); !strings.Contains(body, "ultrasaurus") {
		t.Errorf("Body was not correctly decoded. Should have an ultrasaurus. Instead got:\n\n%v\n\n", body)
	}
}

func TestGmailToMessage(t *testing.T) {
	json := getEmailJson()
	rawGmail, _ := JsonToGmail(json)
	msg, err := GmailToMessage(rawGmail)
	if err != nil {
		t.Errorf("Unexpected error calling JsonForElasticsearch: %v", err)
	}
	if body := msg.Body; !strings.Contains(body, "ultrasaurus") {
		t.Errorf("Body is incorrect. Should have an ultrasaurus. Instead got:\n\n%v\n\n", body)
	}
}

