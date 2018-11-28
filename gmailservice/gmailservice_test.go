package gmailservice

import (
	"testing"
	// "encoding/json"
	"log"
	"io/ioutil"
	"os"
	"strings"
)


func TestBodyText(t *testing.T) {
	jsonFile, err := os.Open("../gmail_message.json") // Slack notification

	if err != nil {
		log.Fatalf("Couldn't open gmail_message.json. Error: %v", err)
	}
	
	// defer jsonFile.Close()

	json, _ := ioutil.ReadAll(jsonFile) // theoretically the same thing that the gmail API returns
	// log.Printf("as a string: %v", string(json))
	jsonFile.Close()
	doc := GmailDoc{source: json}

	if body := doc.BodyText(); !strings.Contains(body, "ultrasajurus") {
		t.Errorf("Body was not correctly decoded. Should have an ultrasaurus. Instead got:\n\n%v\n\n", body)
	}
}

