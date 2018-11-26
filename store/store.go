package store

import (
  "encoding/json"
  "log"
  "github.com/olivere/elastic"
  "golang.org/x/net/context"
)


// Service struct to keep state we need
type Service struct {
  client  *elastic.Client
  ctx     context.Context
}

const IndexName = "mail"

// New returns query.Service initialized with elastic client
func New(ctx context.Context) (*Service, error) {

  client, err := elastic.NewClient();
  if err != nil {
    log.Printf("could not create elastic client, %v", err)
    return nil, err
  }

  exists, err := client.IndexExists(IndexName).Do(ctx);
  if err != nil {
    log.Printf("failed to discover if index '%s' exists, %v", IndexName, err)
    return nil, err
  }

  if !exists {
    if _, err := client.CreateIndex(IndexName).Do(ctx); err != nil {
      log.Printf("failed to create '%s' index, %v", IndexName, err)
      return nil, err
    }
  }

  s := new(Service)
  s.client = client;
  s.ctx = ctx;

  return s, nil
}

// Save in ElasticSearch
func Save(s *Service, byt []byte) (error) {
  var data map[string]interface{}

  if err := json.Unmarshal(byt, &data); err != nil {
    log.Printf("json.Unmarshal failed, skipping meesage, err: ", err)
    return err;
  }
  log.Println("saving Message ID: ", data["id"])
  id := data["id"].(string)

	record, err := s.client.Index().
		Index(IndexName).
    Id(id).
    Type("document").
    BodyJson(string(byt)).
    Do(s.ctx);

  if err != nil {
    if record != nil {
      log.Printf("Failed to index data id %s in index %s, err: %v", record.Id, record.Index, err)
    } else {
      log.Printf("Failed to index. err: %v", err)
    }
    return err;
  }
	log.Printf("Indexed data id %s to index %s, type %s\n", record.Id, record.Index, record.Type)

  return nil;
}
