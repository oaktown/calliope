package main

import (
	"log"
	"calliope/auth"
	"calliope/gmailservice"

	"fmt"
	"golang.org/x/net/context"
  "github.com/olivere/elastic"

)



func main() {
	fmt.Printf("hello, world\n")

	ctx := context.Background()
	client, err := auth.Client(ctx);
	if err != nil {
		log.Fatalf("could not create client, %v", err)
	}
	gsvc, err := gmailservice.New(ctx, client);
	if err != nil {
		log.Fatalf("could not create gmailservice, %v", err)
	} else {
		gmailservice.Download(gsvc);
	}
	log.Printf("done")

}


