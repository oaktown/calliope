package misc

import (
  "context"
  "log"

  "github.com/oaktown/calliope/auth"
  "github.com/oaktown/calliope/store"
  "google.golang.org/api/gmail/v1"
)

var ctx = context.Background()

func GetGmailClient() (*gmail.Service) {
  client, err := auth.Client(ctx)
  if err != nil {
    log.Fatalf("could not get auth client, %v", err)
  }
  svc, err := gmail.New(client)
  if err != nil {
    log.Fatalf("could not create gmail client, %v", err)
  }
  return svc
}

func GetStoreClient() (*store.Service) {
  s, err := store.New(ctx)
  if err != nil {
    log.Fatalf("could not create store, %v", err)
  }
  return s
}
