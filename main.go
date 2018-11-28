package main

import (
	"calliope/auth"
	"calliope/gmailservice"
	"calliope/store"
	"golang.org/x/net/context"
	"log"
	"sync"
)

func reader(s store.Storable, messageChannel <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done() // WaitGroup done when this routines exits

	for byt := range messageChannel { // reads from channel until it's closed
		s.Save(byt)
	}
}

func main() {

	ctx := context.Background()
	client, err := auth.Client(ctx)
	if err != nil {
		log.Fatalf("could not get auth client, %v", err)
	}
	gsvc, err := gmailservice.New(ctx, client)
	if err != nil {
		log.Fatalf("could not create gmailservice, %v", err)
	}

	s, err := store.New(ctx)
	if err != nil {
		log.Fatalf("could not create store, %v", err)
	}

	var wg sync.WaitGroup

	const BufferSize = 10
	messages := make(chan []byte, BufferSize)

	wg.Add(1)
	go reader(s, messages, &wg)

	gmailservice.Download(gsvc, messages)

	wg.Wait()
}
