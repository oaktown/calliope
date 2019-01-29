package store

import (
  "encoding/json"
  "errors"
  "fmt"
  "github.com/olivere/elastic"
  "golang.org/x/net/context"
  "google.golang.org/api/gmail/v1"
  "log"
  "reflect"
  "time"
)

type Service struct {
  Client      *elastic.Client
  Ctx         context.Context
  MailIndex   string
  LabelsIndex string
}

type Message struct {
  Id                  string
  Url                 string
  ThreadId            string
  LabelIds            []string
  Date                time.Time
  DownloadedStartedAt time.Time
  To                  string
  Cc                   string
  From                string
  Subject             string
  Snippet             string
  Body                string
  Source              gmail.Message
}

const MailIndex = "mail"

// New returns Elastic initialized with elastic client
func New(ctx context.Context) (*Service, error) {
  client, err := elastic.NewClient()
  if err != nil {
    log.Println("could not create elastic client: ", err)
    return nil, err
  }

  if err := createIndex(MailIndex, client, ctx); err != nil {
    log.Printf("Error creating %v index: %v\n", MailIndex, err)
    return nil, err
  }
  if err := createIndex(LabelsIndex, client, ctx); err != nil {
    log.Printf("Error creating %v index: %v\n", LabelsIndex, err)
    return nil, err
  }

  svc := Service{
    Client:      client,
    Ctx:         ctx,
    MailIndex:   MailIndex,
    LabelsIndex: LabelsIndex,
  }
  return &svc, nil
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

func (s *Service) saveDoc(index string, id string, json string) (*elastic.IndexResponse, error) {
  response, err := s.Client.Index().
    Index(index).
    Id(id).
    Type("document").
    BodyJson(json).
    Do(s.Ctx)
  if err != nil {
    log.Printf("################ Failed to index data id %s in index %s, err: %v", id, index, err)
  }
  return response, err
}

type MessageResponse struct {
  Message  Message
  Response *elastic.IndexResponse
}

func (s *Service) SaveMessage(data Message, responses chan<- *MessageResponse) error {
  log.Println("saving Message ID: ", data.Id)
  messageJson, _ := json.MarshalIndent(data, "", "\t")
  response, err := s.saveDoc(MailIndex, data.Id, string(messageJson))
  if err != nil {
    return err
  }
  alert := ""
  if response.Version > 1 {
    alert = "***********************"
    responses <- &MessageResponse{
      Message:  data,
      Response: response,
    }
  }
  log.Printf("\nIndexed Message\nid: %s\n%s version:%d\nSubject:%s\n\n", response.Id, alert, response.Version, data.Subject)
  return nil
}

func (s *Service) FindLabelId(labelName string) (string, error) {
  labels, err := s.GetLabels(false)
  if err != nil {
    return "", errors.New("Could not get labels from Elasticsearch")
  }
  var labelId string
  for _, label := range labels {
    if label.Name == labelName {
      labelId = label.Id
    }
  }
  if labelId == "" {
    err := fmt.Sprintf("Label %v not found.", labelName)
    return "", errors.New(err)
  }
  return labelId, nil
}

func (s *Service) GetMessages(req *elastic.SearchService) ([]*Message, error) {
  var messages []*Message
  result, err := req.Do(s.Ctx)
  if err != nil {
    log.Println("Couldn't search. Exiting due to: ", err)
    return nil, err
  }
  var messageForReflect Message
  for _, m := range result.Each(reflect.TypeOf(messageForReflect)) {
    message := m.(Message)
    messages = append(messages, &message)
  }
  log.Println("Messages found: ", len(messages))
  return messages, nil
}

type Stats struct {
  Earliest time.Time
  Latest   time.Time
  Total    int64
}

func (s *Service) GetStats() (Stats, error) {
  var stats Stats
  builder := s.Client.Search().Index(MailIndex).Query(elastic.NewMatchAllQuery())
  builder = builder.Aggregation("maxDate", elastic.NewMaxAggregation().Field("Date"))
  builder = builder.Aggregation("minDate", elastic.NewMinAggregation().Field("Date"))
  results, err := builder.Pretty(true).Do(s.Ctx)
  if err != nil {
    log.Println("Error with getting stats: ", err)
    return stats, err
  }

  stats.Total = results.Hits.TotalHits

  aggs := results.Aggregations

  max, found := aggs.Max("maxDate")
  if found != true {
    log.Println("Could not get maxDate")
  }
  stats.Latest = time.Unix(msToSecs(*max.Value), 0)

  min, found := aggs.Min("minDate")
  if found != true {
    log.Println("Could not get minDate")
  }
  stats.Earliest = time.Unix(msToSecs(*min.Value), 0)

  return stats, nil
}

func msToSecs(timeInMs float64) int64 {
  return int64(timeInMs / 1000.0)
}
