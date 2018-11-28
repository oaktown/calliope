package gmailservice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/gmail/v1"
)

type Message struct {
	Id      string
	Date    time.Time
	To      string
	Cc      string
	Bcc     string
	From    string
	ReplyTo []string // can get multiple reply-to headers in an email
	Subject string
	Body    string // the thing we're decoding
	Source  string // original json
}

// New returns a GmailService value initialized with given http client.
func New(ctx context.Context, client *http.Client) (*gmail.Service, error) {
	gs, err := gmail.New(client)
	if err != nil {
		log.Print("could not create gmail client,", err)
		return nil, err
	}
	return gs, nil
}

// Download doesn't do anything yet
// TODO: Change this to use a channel of type JsonForElasticsearch
// func Download(g *GmailService, messages chan<- JsonForElasticsearch) {
func Download(gs *gmail.Service, lastDate string, pageToken string, batchSize int) ([]Message, error) {
	log.Println("Retrieving messages starting on", lastDate)
	query := gs.Users.Messages.List("me").Q("after: " + lastDate)
	
	if pageToken != "" {
		query.PageToken(pageToken)
	}
	result, err := query.Do()
	if err != nil {
		log.Printf("Unable to retrieve messages: %v", err)
		return nil, err
	}

	log.Printf("Processing %v messages...\n", len(result.Messages))

	var messages []Message
	for _, msgInfo := range result.Messages[:6] {
		msg, err := gs.Users.Messages.Get("me", msgInfo.Id).Do()
		if err != nil {
			log.Printf("Unable to retrieve message %v: %v", msgInfo.Id, err)
			continue
		}

		// DECODING
		{
			// fmt.Printf("Found Message ID: %v\n", msg.Id)
			// data, err := json.MarshalIndent(msg, "", "\t")
			// if err != nil {
			// 	return nil, err
			// }

			// var gm GmailMessage
			// if err := json.Unmarshal(data, &gm); err != nil {
			// 	log.Printf("json.Unmarshal failed, skipping message, err: %v", err)
			// 	return nil, err
			// }

			// TODO: Use a JsonForElasticsearch instead of bytestream
			// something like:
			// doc := GmailDoc{source: byt}
			// messages <- doc.JsonForElasticsearch()
		}

		var m Message
		messages = append(messages, m)
	}
	return messages, nil
}

type GmailDoc struct {
	source []byte
}

type GmailMessage struct {
	HistoryId    string   `json:"historyId"`
	Id           string   `json:"id"`
	InternalDate string   `json:"internalDate"`
	LabelIds     []string `json:labelIds`
	Payload      GmailMessagePayload
}

type GmailMessagePayload struct {
	Parts []GmailMessagePart `json:"parts"`
}

type GmailMessagePart struct {
	Body struct {
		Data string `json:"data"`
	} `json:"body"`
	MimeType string `json:"mimeType"`
}

func (doc *GmailDoc) JsonData() (GmailMessage, error) {
	var data GmailMessage
	if err := json.Unmarshal(doc.source, &data); err != nil {
		log.Printf("json.Unmarshal failed, skipping message, err: %v", err)
		return data, err
	}
	return data, nil
}

func (doc *GmailDoc) BodyText() string {
	data, err := doc.JsonData()
	if err != nil {
		return ""
	}
	parts := data.Payload.Parts
	for _, part := range parts {
		if part.MimeType == "text/plain" {
			encodedBody := part.Body.Data
			// log.Printf("body: %v", encodedBody)
			body, _ := base64.URLEncoding.DecodeString(encodedBody)
			return string(body)
		}
	}
	//	doc.source = "" // TODO: Figure out golang thing (nothing to do with this method). Can we mutate ourself?
	return "" // TODO: is this the right thing to do when not found? Possibly should look at body field?
}

func (doc *GmailDoc) JsonForElasticsearch() (JsonForElasticsearch, error) {
	jsonStruct := JsonForElasticsearch{
		Body: (doc.BodyText()),
	}

	return jsonStruct, nil

	// returns json in the format we want to save in Elasticsearch
	// return JsonForElasticsearch{
	// 	Id doc.
	// 	Date Time
	// 	To string
	// 	Cc string
	// 	Bcc string
	// 	From string
	// 	ReplyTo []string // can get multiple reply-to headers in an email
	// 	Subject string
	// 	Body string // the thing we're decoding
	// 	Source string // original json
	// }
}
