package airport

import (
	"context"
	"log"
	"strings"

	db "bilecik/pkg"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

const SearchLimit = 6

type Repository struct {
	db *db.DB
	es *elasticsearch.Client
}

// NewRepository creates the airport repository; es may be nil, then search
// runs on Postgres ILIKE instead of Elasticsearch.
func NewRepository(db *db.DB, es *elasticsearch.Client) *Repository {
	return &Repository{
		db: db,
		es: es,
	}
}

const searchQuery = `
SELECT id, name, city, country, iata_code, icao_code, language, popularity
FROM (
    SELECT DISTINCT ON (iata_code) id, name, city, country, iata_code, icao_code, language, popularity, relevance
    FROM (
        SELECT id, name, city, country, iata_code, icao_code, language, popularity,
            CASE
                WHEN iata_code ILIKE @q THEN 1
                WHEN icao_code ILIKE @q THEN 2
                WHEN city ILIKE @q THEN 3
                WHEN country ILIKE @q THEN 4
                WHEN name ILIKE @q THEN 5
                WHEN city ILIKE '%' || @q || '%' THEN 6
                WHEN country ILIKE '%' || @q || '%' THEN 7
                ELSE 8
            END AS relevance
        FROM airports
        WHERE iata_code ILIKE '%' || @q || '%'
            OR icao_code ILIKE '%' || @q || '%'
            OR city ILIKE '%' || @q || '%'
            OR country ILIKE '%' || @q || '%'
            OR name ILIKE '%' || @q || '%'
    ) AS matched
    ORDER BY iata_code, relevance
) AS deduped
ORDER BY relevance, popularity DESC, name
LIMIT @limit
`

func (repo *Repository) Search(ctx context.Context, query string) ([]Airport, error) {
	query = strings.TrimSpace(query)
	if repo.es != nil {
		airports, err := repo.searchElastic(ctx, query)
		if err == nil {
			return airports, nil
		}
		log.Printf("elastic search failed, falling back to postgres: %v", err)
	}
	return repo.searchSQL(ctx, query)
}

func (repo *Repository) searchSQL(ctx context.Context, query string) ([]Airport, error) {
	var airports []Airport
	err := repo.db.WithContext(ctx).
		Raw(searchQuery, map[string]any{
			"q":     escapeLike(query),
			"limit": SearchLimit,
		}).
		Scan(&airports).Error
	return airports, err
}

func (repo *Repository) GetByIATA(ctx context.Context, iata string) (*Airport, error) {
	var airports []Airport
	err := repo.db.WithContext(ctx).
		Raw(`
			SELECT id, name, city, country, iata_code, icao_code, language
			FROM airports
			WHERE iata_code = @iata
			ORDER BY CASE WHEN language = 'ru' THEN 0 ELSE 1 END
			LIMIT 1
		`, map[string]any{"iata": strings.ToUpper(strings.TrimSpace(iata))}).
		Scan(&airports).Error
	if err != nil || len(airports) == 0 {
		return nil, err
	}
	return &airports[0], nil
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
