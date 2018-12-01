package misc

import (
  "context"
  "log"

  "github.com/oaktown/calliope/auth"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/oaktown/calliope/store"
  "google.golang.org/api/gmail/v1"
)

var ctx = context.Background()

func GetGmailClient() (*gmail.Service) {
  client, err := auth.Client(ctx)
  if err != nil {
    log.Fatalf("could not get auth client, %v", err)
  }
  gsvc, err := gmailservice.New(client)
  if err != nil {
    log.Fatalf("could not create gmailservice, %v", err)
  }
  return gsvc
}

func GetStoreClient() (*store.Service) {
  s, err := store.New(ctx)
  if err != nil {
    log.Fatalf("could not create store, %v", err)
  }
  return s
}
