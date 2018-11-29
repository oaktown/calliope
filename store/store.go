package store

import (
  "encoding/json"
  "log"
  "github.com/olivere/elastic"
	"golang.org/x/net/context"
	"github.com/oaktown/calliope/gmailservice"
)

type Storable interface {
  Save(gmailservice.Message) error
}

// Service struct to keep state we need
type Service struct {
  client  *elastic.Client
  ctx     context.Context
}

const IndexName = "mail"

// New returns Elastic initialized with elastic client
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

func (s *Service) Save(data gmailservice.Message) error {
  // var data map[string]interface{}

  // if err := json.Unmarshal(byt, &data); err != nil {
  //   log.Printf("json.Unmarshal failed, skipping meesage, err: ", err)
  //   return err;
  // }
  log.Println("saving Message ID: ", data.Id)
	json, err := json.MarshalIndent(data, "", "\t")
	
	record, err := s.client.Index().
		Index(IndexName).
    Id(data.Id).
    Type("document").
    BodyJson(string(json)).
    Do(s.ctx);

  if err != nil {
		log.Printf("Failed to index data id %s in index %s, err: %v", data.Id, IndexName, err)
    return err;
  }
	log.Printf("Indexed data id %s to index %s, type %s\n", record.Id, record.Index, record.Type)

  return nil;
}
