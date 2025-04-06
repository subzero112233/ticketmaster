package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/subzero112233/ticketmaster/domain/entity"
	"github.com/subzero112233/ticketmaster/usecase/events"
	"io"
	"time"
)

type ElasticsearchRepository struct {
	esClient  *elasticsearch.Client
	indexName string
}

func NewElasticSearchImplementation(esClient *elasticsearch.Client, indexName string) *ElasticsearchRepository { // nolint
	return &ElasticsearchRepository{
		esClient:  esClient,
		indexName: indexName,
	}
}

func (es *ElasticsearchRepository) SearchEvents(ctx context.Context, filter *events.Filter) ([]entity.Event, error) {
	query := es.buildSearchQuery(filter)

	// Execute the search
	res, err := es.esClient.Search(
		es.esClient.Search.WithContext(ctx),
		es.esClient.Search.WithIndex(es.indexName),
		es.esClient.Search.WithBody(query),
		es.esClient.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	var searchResponse SearchResult
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Extract output from the search response
	output := make([]Event, len(searchResponse.Hits.Hits))
	for i, hit := range searchResponse.Hits.Hits {
		output[i] = hit.Source
	}

	return toEntityEvents(output), nil
}

func (es *ElasticsearchRepository) buildSearchQuery(filter *events.Filter) io.Reader {
	var mustClauses []interface{}
	var filterClauses []interface{}

	if filter.Description != nil && *filter.Description != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match_phrase": map[string]interface{}{
				"description": *filter.Description, // Elasticsearch will handle multi-word phrases
			},
		})
	}

	if filter.Location != nil && *filter.Location != "" {
		filterClauses = append(filterClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"location": *filter.Location,
			},
		})
	}

	if !filter.FromDate.IsZero() && !filter.ToDate.IsZero() {
		fromDays := dateToDaysSinceEpoch(filter.FromDate)
		toDays := dateToDaysSinceEpoch(filter.ToDate)

		filterClauses = append(filterClauses, map[string]interface{}{
			"range": map[string]interface{}{
				"date": map[string]interface{}{
					"gte": fromDays,
					"lte": toDays,
				},
			},
		})
	}

	query := map[string]interface{}{
		"from": (filter.Page - 1) * 10,
		"size": 10,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   mustClauses,
				"filter": filterClauses,
			},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		fmt.Printf("failed to marshal query: %v\n", err)
		return nil
	}

	return bytes.NewReader(queryBytes)
}

func dateToDaysSinceEpoch(t time.Time) int {
	epoch := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	return int(t.Sub(epoch).Hours() / 24)
}

func toEntityEvents(events []Event) []entity.Event {
	output := make([]entity.Event, len(events))
	for i, event := range events {
		output[i] = toEntityEvent(event)
	}

	return output
}

func toEntityEvent(event Event) entity.Event {
	return entity.Event{
		Date:        time.Unix(event.Date, 0),
		ID:          event.ID,
		Location:    event.Location,
		Name:        event.Name,
		Performer:   event.Performer,
		Venue:       event.Venue,
		Description: event.Description,
	}
}

type Event struct {
	Date        int64  `json:"date"` // date is stored as an integer in elasticsearch
	Venue       string `json:"venue"`
	Performer   string `json:"performer"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	ID          string `json:"id"`
}

type SearchResult struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Index  string  `json:"_index"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source Event   `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
