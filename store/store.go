package store

import (
  "encoding/json"
  "errors"
  "fmt"
  "github.com/oaktown/calliope/gmailservice"
  "github.com/olivere/elastic"
  "golang.org/x/net/context"
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

func (s *Service) saveDoc(index string, id string, json string) error {
  record, err := s.Client.Index().
    Index(index).
    Id(id).
    Type("document").
    BodyJson(json).
    Do(s.Ctx)
  if err != nil {
    log.Printf("Failed to index data id %s in index %s, err: %v", id, index, err)
    return err
  }
  log.Printf("Indexed data id %s to index %s, type %s\n", id, record.Index, record.Type)
  return nil
}

type LabelsDoc struct {
  Id     string
  Labels []*gmailservice.Label
}

func (s *Service) SaveLabels(labels []*gmailservice.Label) error {
  doc := LabelsDoc{
    Id:     "labels",
    Labels: labels,
  }
  labelsJson, _ := json.MarshalIndent(doc, "", "\t")

  if err := s.saveDoc(LabelsIndex, "labels", string(labelsJson)); err != nil {
    return err
  }

  return nil
}

func (s *Service) GetLabels() ([]*gmailservice.Label, error) {
  var doc LabelsDoc
  query := elastic.NewTermQuery("Id", "labels")
  result, _ := s.Client.Search().
    Index(LabelsIndex).
    Query(query).
    Do(s.Ctx)
  labelsJson := result.Hits.Hits[0].Source
  if err := json.Unmarshal(*labelsJson, &doc); err != nil {
    log.Println("Unable to unmarshal labels json. err: ", err)
    return nil, err
  }
  return doc.Labels, nil
}

func (s *Service) SaveMessage(data gmailservice.Message) error {
  log.Println("saving Message ID: ", data.Id)
  messageJson, _ := json.MarshalIndent(data, "", "\t")
  if err := s.saveDoc(MailIndex, data.Id, string(messageJson)); err != nil {
    return err
  }
  return nil
}

func (s *Service) GenerateMessagesQuery(labelName string, starred bool) (*elastic.BoolQuery, error) {
  labelId, err := s.FindLabelId(labelName)
  if err != nil {
    return nil, err
  }
  labelQuery := elastic.NewTermQuery("LabelIds.keyword", labelId)

  query := elastic.NewBoolQuery()
  query = query.Must(labelQuery) // weirdly before this line was in, when we got source, it included the label query but didn't work
  if starred {
    starredQuery := elastic.NewTermQuery("LabelIds.keyword", "STARRED")
    query = query.Must(starredQuery)
  }
  source, _ := query.Source()
  jsonStr, _ := json.MarshalIndent(source, "", "\t")
  log.Printf("source: %v\n", string(jsonStr))
  return query, nil
}

func (s *Service) FindLabelId(labelName string) (string, error) {
  labels, err := s.GetLabels()
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

func (s *Service) GetMessages(label string, starred bool, size int) ([]*gmailservice.Message, error) {
  query, err := s.GenerateMessagesQuery(label, starred)
  if err != nil {
    log.Println("Query error: ", err)
    return nil, err
  }
  var messages []*gmailservice.Message
  result, err := s.Client.Search().
    Index(s.MailIndex).
    Query(query).
    Size(size).
    Do(s.Ctx)
  if err != nil {
    log.Println("Couldn't search. Exiting")
    return nil, err
  }
  var messageForReflect gmailservice.Message
  for _, m := range result.Each(reflect.TypeOf(messageForReflect)) {
    message := m.(gmailservice.Message)
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
  stats.Latest = time.Unix(toSeconds(*max.Value), 0)

  min, found := aggs.Min("minDate")
  if found != true {
    log.Println("Could not get minDate")
  }
  stats.Earliest = time.Unix(toSeconds(*min.Value), 0)

  return stats, nil
}

func toSeconds(timeInMs float64) int64 {
  return int64(timeInMs / 1000.0)
}
