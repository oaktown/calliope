package store

import (
	"encoding/json"
	"github.com/olivere/elastic"
	"log"
)

type Label struct {
	Id   string
	Name string
}

const LabelsIndex = "labels"

type LabelsDoc struct {
	Id     string
	Labels []*Label
}

func (s *Service) SaveLabels(labels []*Label) error {
	doc := LabelsDoc{
		Id:     "labels",
		Labels: labels,
	}
	labelsJson, _ := json.MarshalIndent(doc, "", "\t")

	if _, err := s.saveDoc(LabelsIndex, "labels", string(labelsJson)); err != nil {
		return err
	}

	return nil
}

func (s *Service) getLabelsFromStore() ([]*Label, error) {
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

var googleLabel = map[string]bool{
	"CATEGORY_PERSONAL":   true,
	"IMPORTANT":           true,
	"CHAT":                true,
	"SENT":                true,
	"INBOX":               true,
	"TRASH":               true,
	"DRAFT":               true,
	"SPAM":                true,
	"STARRED":             true,
	"UNREAD":              true,
	"CATEGORY_FORUMS":     true,
	"CATEGORY_SOCIAL":     true,
	"CATEGORY_UPDATES":    true,
	"CATEGORY_PROMOTIONS": true,
	"[Imap]/Drafts":       true,
	"[Imap]/Archive":      true,
	"Deleted Messages":    true,
}

func (s *Service) GetLabels(userOnly bool) ([]*Label, error) {
	labels, err := s.getLabelsFromStore()
	if err != nil {
		// TODO: handle error
	}
	if userOnly {
		userLabels := labels[:0]
		for _, label := range labels {
			if !googleLabel[label.Name] {
				userLabels = append(userLabels, label)
			}
		}
		return userLabels, nil
	}
	return labels, nil
}
