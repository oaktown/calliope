package store

import (
  "encoding/json"
  "github.com/olivere/elastic"
  "strings"
  "time"
)

type MessageSearch interface {
  QueryString() string
  Do() ([]*Message, error)
}

type RawMessageSearch struct {
  svc           *Service
  rawQuery      string
  searchService *elastic.SearchService
}

func (s *Service) NewRawMessageSearch(q string) RawMessageSearch {
  return RawMessageSearch{
    svc:           s,
    rawQuery:      q,
    searchService: s.Client.Search().Source(q),
  }
}

func (r RawMessageSearch) QueryString() string {
  return r.rawQuery
}

func (r RawMessageSearch) Do() ([]*Message, error) {
  return r.svc.GetMessages(r.searchService)
}

type StructuredMessageSearch struct {
  svc           *Service
  query         *elastic.BoolQuery
  searchSource  *elastic.SearchSource
  searchService *elastic.SearchService
}

func (s *Service) NewStructuredMessageSearch() StructuredMessageSearch {
  return StructuredMessageSearch{
    svc: s,
  }
}

func (s StructuredMessageSearch) newOrExistingQuery() *elastic.BoolQuery {
  var query *elastic.BoolQuery
  if s.query == nil {
    query = elastic.NewBoolQuery()
  } else {
    query = s.query
  }
  return query
}

func (s StructuredMessageSearch) Label(labelName string) StructuredMessageSearch {
  labelId, err := s.svc.FindLabelId(labelName)
  if err != nil {
    // TODO: Deal with errors. Maybe add an errors field; in the meantime, do nothing
    return s
  }
  labelQuery := elastic.NewTermQuery("LabelIds.keyword", labelId)
  query := s.newOrExistingQuery()
  s.query = query.Must(labelQuery)
  return s
}

func (s StructuredMessageSearch) Starred(starred bool) StructuredMessageSearch {
  if !starred {
    return s
  }
  query := s.newOrExistingQuery()
  starredQuery := elastic.NewTermQuery("LabelIds.keyword", "STARRED")
  s.query = query.Must(starredQuery)
  return s
}

func (s StructuredMessageSearch) Participants(participants string) StructuredMessageSearch {
  if participants == "" {
    return s
  }
  emails := strings.Split(participants, ",")
  query := s.newOrExistingQuery()
  for _, email := range emails {
    mm := elastic.NewMultiMatchQuery(email, "From", "To", "Cc").Type("cross_fields").Operator("and")
    query = query.Must(mm)
  }
  s.query = query
  return s
}

func (s StructuredMessageSearch) DateRange(d1, d2 string) StructuredMessageSearch {
  query := s.newOrExistingQuery()
  startDate, startErr := time.Parse("2006-01-02", d1)
  endDate, endErr := time.Parse("2006-01-02", d2)
  if startErr != nil && endErr != nil {
    // No valid date strings were passed in (includes case of two empty strings)
    return s
  }

  rangeQuery := elastic.NewRangeQuery("Date")

  if startErr == nil {
    rangeQuery.Gte(startDate)
  }

  if endErr == nil {
    // Add a day to account for hours after midnight
    rangeQuery.Lte(endDate.AddDate(0,0,1))
  }

  s.query = query.Must(rangeQuery)

  return s
}

func (s StructuredMessageSearch) EndDate(endDate string) StructuredMessageSearch {
  return s
}

func (s StructuredMessageSearch) Size(size int) StructuredMessageSearch {
  searchSource := s.getSearchSource()
  s.searchSource = searchSource.Size(size)
  s.searchService = s.svc.Client.Search().SearchSource(s.searchSource)
  return s
}

func (s StructuredMessageSearch) Sort(field string, asc bool) StructuredMessageSearch {
  searchSource := s.getSearchSource()
  s.searchSource = searchSource.Sort(field, asc)
  s.searchService = s.svc.Client.Search().SearchSource(s.searchSource)
  return s
}

func (s StructuredMessageSearch) QueryString() string {
  source, _ := s.getSearchSource().Source()
  queryJson, _ := json.MarshalIndent(source, "", "  ")
  query := string(queryJson)
  return query
}

func (s StructuredMessageSearch) Do() ([]*Message, error) {
  return s.svc.GetMessages(s.searchService)
}

func (s StructuredMessageSearch) getSearchSource() *elastic.SearchSource {
  if s.searchSource != nil {
    return s.searchSource
  }
  var query elastic.Query
  if s.query == nil {
    query = elastic.NewMatchAllQuery()
  } else {
    query = s.query
  }
  searchSource := elastic.NewSearchSource().Query(query)
  return searchSource
}
