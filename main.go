package main

import (
	"github.com/oaktown/calliope/auth"
	"github.com/oaktown/calliope/gmailservice"
	"github.com/oaktown/calliope/store"
	"golang.org/x/net/context"
	"log"
	"os"
	"sync"
)

func reader(s store.Storable, messageChannel <-chan gmailservice.Message, wg *sync.WaitGroup) {
	defer wg.Done() // WaitGroup done when this routines exits

	for message := range messageChannel { // reads from channel until it's closed
		s.Save(message)
	}
}

func main() {

	ctx := context.Background()
	client, err := auth.Client(ctx)
	if err != nil {
		log.Fatalf("could not get auth client, %v", err)
	}
	gsvc, err := gmailservice.New(client)
	if err != nil {
		log.Fatalf("could not create gmailservice, %v", err)
	}

	s, err := store.New(ctx)
	if err != nil {
		log.Fatalf("could not create store, %v", err)
	}

	var wg sync.WaitGroup

	const BufferSize = 10
	messageChannel := make(chan gmailservice.Message, BufferSize)

	wg.Add(1)
	go reader(s, messageChannel, &wg)

	var inboxUrl string
	if os.Getenv("CALLIOPE_INBOX_URL") == "" {
		inboxUrl = "https://mail.google.com/mail/#inbox"
	} else {
		inboxUrl = os.Getenv("CALLIOPE_INBOX_URL")
	}
	messages, err := gmailservice.Download(gsvc, "2018/01/01", 6, "", inboxUrl)
	if err != nil {
		log.Fatal("Unable to download messageChannel. Error: ", err)
	}

	for _, message := range messages {
		messageChannel <- message
	}

	close(messageChannel)

	wg.Wait()
}
