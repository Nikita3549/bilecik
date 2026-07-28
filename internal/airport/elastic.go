package airport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	db "bilecik/pkg"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const esIndex = "airports"

const esIndexSettings = `{
	"analysis": {
		"filter": {
			"autocomplete_filter": { "type": "edge_ngram", "min_gram": 1, "max_gram": 20 }
		},
		"analyzer": {
			"autocomplete": {
				"type": "custom",
				"tokenizer": "standard",
				"filter": ["lowercase", "autocomplete_filter"]
			}
		}
	}
}`

const esIndexMappings = `{
	"properties": {
		"name": {
			"type": "text",
			"analyzer": "standard",
			"fields": { "keyword": { "type": "keyword", "ignore_above": 256 } }
		},
		"city": {
			"type": "text",
			"fields": { "keyword": { "type": "keyword", "ignore_above": 256 } }
		},
		"country": {
			"type": "text",
			"fields": { "keyword": { "type": "keyword", "ignore_above": 256 } }
		},
		"iata_code": { "type": "keyword" },
		"icao_code": { "type": "keyword" },
		"language": { "type": "keyword" },
		"popularity": { "type": "integer" }
	}
}`

var likelyCodeRe = regexp.MustCompile(`^[A-Za-z]{3,4}$`)

func NewElasticClient(url string) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{url}})
	if err != nil {
		return nil, fmt.Errorf("elastic client: %w", err)
	}
	return es, nil
}

func SyncElastic(ctx context.Context, es *elasticsearch.Client, database *db.DB) error {
	if err := waitForElastic(ctx, es, 2*time.Minute); err != nil {
		return err
	}

	if err := recreateIndex(ctx, es); err != nil {
		return err
	}

	var airports []Airport
	if err := database.WithContext(ctx).Find(&airports).Error; err != nil {
		return fmt.Errorf("load airports from db: %w", err)
	}
	if len(airports) == 0 {
		return fmt.Errorf("airports table is empty, nothing to index")
	}

	var buf bytes.Buffer
	for _, a := range airports {
		meta := fmt.Sprintf(`{"index":{"_index":%q,"_id":"%d_%s_%s"}}`, esIndex, a.ID, a.IATACode, a.Language)
		doc, err := json.Marshal(a)
		if err != nil {
			return fmt.Errorf("marshal airport doc: %w", err)
		}
		buf.WriteString(meta)
		buf.WriteByte('\n')
		buf.Write(doc)
		buf.WriteByte('\n')
	}

	res, err := es.Bulk(&buf, es.Bulk.WithContext(ctx), es.Bulk.WithRefresh("true"))
	if err != nil {
		return fmt.Errorf("bulk index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("bulk index: %s", resBody(res))
	}

	var bulkRes struct {
		Errors bool `json:"errors"`
	}
	if err := json.NewDecoder(res.Body).Decode(&bulkRes); err != nil {
		return fmt.Errorf("decode bulk response: %w", err)
	}
	if bulkRes.Errors {
		return fmt.Errorf("bulk index finished with item errors")
	}
	return nil
}

func waitForElastic(ctx context.Context, es *elasticsearch.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		res, err := es.Ping(es.Ping.WithContext(ctx))
		if err == nil && !res.IsError() {
			res.Body.Close()
			return nil
		}
		if res != nil {
			res.Body.Close()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("elastic not reachable after %s: %w", timeout, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

func recreateIndex(ctx context.Context, es *elasticsearch.Client) error {
	exists, err := es.Indices.Exists([]string{esIndex}, es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("index exists check: %w", err)
	}
	exists.Body.Close()
	if exists.StatusCode == 200 {
		del, err := es.Indices.Delete([]string{esIndex}, es.Indices.Delete.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("delete index: %w", err)
		}
		del.Body.Close()
	}

	body := fmt.Sprintf(`{"settings":%s,"mappings":%s}`, esIndexSettings, esIndexMappings)
	res, err := es.Indices.Create(
		esIndex,
		es.Indices.Create.WithContext(ctx),
		es.Indices.Create.WithBody(strings.NewReader(body)),
	)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("create index: %s", resBody(res))
	}
	return nil
}

func (repo *Repository) searchElastic(ctx context.Context, query string) ([]Airport, error) {
	queryUpper := strings.ToUpper(query)
	isLikelyCode := likelyCodeRe.MatchString(query)

	var should []map[string]any
	if isLikelyCode {
		should = append(
			should,
			map[string]any{"term": map[string]any{"iata_code": map[string]any{"value": queryUpper, "boost": 10000}}},
			map[string]any{"term": map[string]any{"icao_code": map[string]any{"value": queryUpper, "boost": 10000}}},
		)
	}
	should = append(
		should,
		map[string]any{"term": map[string]any{"country.keyword": map[string]any{"value": query, "case_insensitive": true, "boost": 5000}}},
		map[string]any{"term": map[string]any{"city.keyword": map[string]any{"value": query, "case_insensitive": true, "boost": 4000}}},
		map[string]any{"multi_match": map[string]any{
			"query":    query,
			"type":     "cross_fields",
			"fields":   []string{"city^3", "name^2", "country"},
			"operator": "and",
			"boost":    2000,
		}},
		map[string]any{"multi_match": map[string]any{
			"query":  query,
			"type":   "phrase_prefix",
			"fields": []string{"city", "name", "country"},
			"boost":  1000,
		}},
		map[string]any{"multi_match": map[string]any{
			"query":         query,
			"fields":        []string{"name", "city"},
			"fuzziness":     "AUTO",
			"prefix_length": 1,
			"boost":         10,
		}},
	)

	filter := []map[string]any{}
	if isLikelyCode {
		filter = append(filter, map[string]any{"term": map[string]any{"language": "en"}})
	}

	body := map[string]any{
		"size":     SearchLimit,
		"collapse": map[string]any{"field": "iata_code"},
		"query": map[string]any{
			"function_score": map[string]any{
				"query": map[string]any{
					"bool": map[string]any{
						"should":               should,
						"minimum_should_match": 1,
						"filter":               filter,
					},
				},
				"functions": []map[string]any{
					{"field_value_factor": map[string]any{
						"field":    "popularity",
						"modifier": "log2p",
						"factor":   1.5,
						"missing":  1,
					}},
				},
				"boost_mode": "multiply",
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal search query: %w", err)
	}

	res, err := repo.es.Search(
		repo.es.Search.WithContext(ctx),
		repo.es.Search.WithIndex(esIndex),
		repo.es.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic search: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("elastic search: %s", resBody(res))
	}

	var parsed struct {
		Hits struct {
			Hits []struct {
				Source Airport `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	airports := make([]Airport, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		airports = append(airports, hit.Source)
	}
	return airports, nil
}

func resBody(res *esapi.Response) string {
	b, _ := io.ReadAll(res.Body)
	return res.Status() + " " + string(b)
}
