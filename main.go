package main

import (
  "fmt"
  "log"
  "calliope/auth"
  "calliope/gmailservice"
  "encoding/json"
  "sync"
  "golang.org/x/net/context"
)

func reader(messageChannel <-chan []byte, wg *sync.WaitGroup) {
  var data map[string]interface{}

  defer wg.Done()		// WaitGroup done when this routines exits

  for byt := range messageChannel { // reads from channel until it's closed
    if err := json.Unmarshal(byt, &data); err != nil {
      log.Printf("json.Unmarshal failed, skipping meesage, err: ", err)
    }
    fmt.Println("recieved Message ID: ", data["id"])
  }
}

func main() {

  ctx := context.Background()
  client, err := auth.Client(ctx);
  if err != nil {
    log.Fatalf("could not create client, %v", err)
  }
  gsvc, err := gmailservice.New(ctx, client);
  if err != nil {
    log.Fatalf("could not create gmailservice, %v", err)
  }

  var wg sync.WaitGroup

  const BufferSize = 10;
  messages := make(chan []byte, BufferSize)

  wg.Add(1)
  go reader(messages, &wg)

  gmailservice.Download(gsvc, messages)

  wg.Wait()
}


