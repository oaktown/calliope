package store

import (
  "encoding/json"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/olivere/elastic"
  "golang.org/x/net/context"
  "log"
)

type Storable interface {
  SaveMessage(gmailservice.Message) error
}

// Service struct to keep state we need
type Service struct {
  client *elastic.Client
  ctx    context.Context
}

const MailIndex = "mail"
const LabelsIndex = "labels"

// New returns Elastic initialized with elastic client
func New(ctx context.Context) (*Service, error) {

  client, err := elastic.NewClient()
  if err != nil {
    log.Println("could not create elastic client: ", err)
    return nil, err
  }

  if err := createIndex(MailIndex, client, ctx); err != nil {
    log.Println("Error creating %v index: %v", MailIndex, err)
    return nil, err
  }
  if err := createIndex(LabelsIndex, client, ctx); err != nil {
    log.Println("Error creating %v index: %v", LabelsIndex, err)
    return nil, err
  }

  svc := new(Service)
  svc.client = client
  svc.ctx = ctx
  return svc, nil
}

func createIndex(name string, client *elastic.Client, ctx context.Context) error {
  exists, err := client.IndexExists(name).Do(ctx)
  if err != nil {
    log.Printf("failed to discover if index '%s' exists, %v", name, err)
    return err
  }
  if !exists {
    if _, err := client.CreateIndex(name).Do(ctx); err != nil {
      log.Printf("failed to create '%s' index, %v", name, err)
      return err
    }
  }
  return nil
}

func (s *Service) saveDoc(index string, id string, json string) error {
  record, err := s.client.Index().
    Index(index).
    Id(id).
    Type("document").
    BodyJson(json).
    Do(s.ctx)
  if err != nil {
    log.Printf("Failed to index data id %s in index %s, err: %v", id, index, err)
    return err
  }
  log.Printf("Indexed data id %s to index %s, type %s\n", id, record.Index, record.Type)
  return nil
}

type LabelsDoc struct {
  Id string
  Labels []*gmailservice.Label
}
func (s *Service) SaveLabels(labels []*gmailservice.Label) error {
  doc := LabelsDoc{
    Id:      "labels",
    Labels : labels,
  }
  labelsJson, _ := json.MarshalIndent(doc, "", "\t")

  if err := s.saveDoc(LabelsIndex, "labels", string(labelsJson)); err != nil {
    return err
  }

  return nil
}

func (s *Service) SaveMessage(data gmailservice.Message) error {
  log.Println("saving Message ID: ", data.Id)
  messageJson, _ := json.MarshalIndent(data, "", "\t")
  if err := s.saveDoc(MailIndex, data.Id, string(messageJson)); err != nil {
    return err
  }
  return nil
}

