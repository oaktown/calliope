package gmailservice

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"google.golang.org/api/gmail/v1"

)


// GmailService keeps state we need
type GmailService struct {
	svc         *gmail.Service
}


// New returns GmailService initialized with given client
func New(ctx context.Context, client *http.Client) (*GmailService, error) {
	m := new(GmailService)
	svc, err := gmail.New(client)
	if err != nil {
		log.Printf("could not create gmail client, %v", err)
		return nil, err
	}
	m.svc = svc;
	return m, nil
}

// Download doesn't do anything yet
func Download(*GmailService) {
	fmt.Printf("ready to do something!\n")

}