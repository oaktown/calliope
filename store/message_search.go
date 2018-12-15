package store

import (
  "encoding/json"
  "github.com/olivere/elastic"
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
  s.query  = query.Must(labelQuery)
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
  return s
}

func (s StructuredMessageSearch) StartDate(startDate string) StructuredMessageSearch {
  return s
}

func (s StructuredMessageSearch) EndDate(endDate string) StructuredMessageSearch {
  return s
}

func (s StructuredMessageSearch) RawQuery(raw string) StructuredMessageSearch {
  return s
}

func (s StructuredMessageSearch) Size(size int) StructuredMessageSearch {
  // In addition to setting size, this is the method that sets searchSource and searchService
  // TODO: This is the last thing called, so here we can determine if there is a search source; if not, then use match all query
  // TODO:  size should be called on a different type than the others (to enforce order of callling; this needs to be last).
  var query elastic.Query
  if s.query == nil {
    query = elastic.NewMatchAllQuery()
  } else {
    query = s.query
  }
  s.searchSource = elastic.NewSearchSource().Query(query).Size(size)
  s.searchService = s.svc.Client.Search().SearchSource(s.searchSource)
  return s
}

func (s StructuredMessageSearch) QueryString() string {
  source, _ := s.searchSource.Source()
  queryJson, _ := json.MarshalIndent(source, "", "  ")
  query := string(queryJson)
  return query
}

func (s StructuredMessageSearch) Do() ([]*Message, error) {
  return s.svc.GetMessages(s.searchService)
}