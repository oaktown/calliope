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

func TestBodyText(t *testing.T) {
	doc := GmailDoc{source: getEmailJson()}

	if body := doc.BodyText(); !strings.Contains(body, "ultrasaurus") {
		t.Errorf("Body was not correctly decoded. Should have an ultrasaurus. Instead got:\n\n%v\n\n", body)
	}
}

func TestJsonForElasticsearch(t *testing.T) {
	doc := GmailDoc{source: getEmailJson()}

	jsonStruct, err := doc.JsonForElasticsearch()
	if err != nil {
		t.Errorf("Unexpected error calling JsonForElasticsearch: %v", err)
	}
	if body := jsonStruct.Body; !strings.Contains(body, "ultrasaurus") {
		t.Errorf("Body is incorrect. Should have an ultrasaurus. Instead got:\n\n%v\n\n", body)
	}
}
