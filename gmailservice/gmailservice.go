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
func Download(g *GmailService) {
	fmt.Printf("ready to do something!\n")

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
		for _, m := range r.Messages[:2] {

			msg, err := g.svc.Users.Messages.Get("me", m.Id).Do()
			if err != nil {
				log.Printf("Unable to retrieve message %v: %v", m.Id, err)
				continue
			}
			fmt.Printf("Message ID: %v\n", m.Id)
			fmt.Printf("%v\n\n", msg)

			// lastMessageRetrievedDate, err := utils.MsToTime(strconv.FormatInt(msg.InternalDate, 10))
			// if err != nil {
			// 	log.Println("Unable to parse message date", err)
			// }

			// //message date
			// log.Println(lastMessageRetrievedDate)

			// if firstMessage {

			// 	//set the last known date
			// 	currentDate := lastMessageRetrievedDate.Format("2006/01/02")
			// 	err = g.db.SetLastDate(currentDate)
			// 	if err != nil {
			// 		log.Println("Unable to save last message date:", currentDate)
			// 	} else {
			// 		log.Println("Saved last message date:", currentDate)
			// 		firstMessage = false
			// 	}

			// }


		if r.NextPageToken == "" {
			break
		}
		pageToken = r.NextPageToken

		//break


	}



}




