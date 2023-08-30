package example

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	paginationaggregator "github.com/Mhakimamransyah/go-pagination-aggregate/pagination_aggregator"
)

type Pokemon struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type PokemonResponse struct {
	Count    int       `json:"count"`
	Pokemons []Pokemon `json:"results"`
}

func (obj PokemonResponse) GetBoundary() int {
	return obj.Count
}

func TestGetPokemon(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	pag, err := paginationaggregator.NewPaginationAggregatorWithContext(ctx, &paginationaggregator.PaginationAggregatorConfig{
		URL:        "https://pokeapi.co/api/v2/pokemon?limit=10&offset=%d",
		JsonPage:   &PokemonResponse{},
		Client:     &http.Client{},
		Concurrent: 5,
		ConcurrentBatchWithContext: func(ctx context.Context, batchResult []paginationaggregator.HttpInteraction) error {

			for _, val := range batchResult {

				tmpData := PokemonResponse{}

				if err := json.Unmarshal([]byte(val.Response.Data), &tmpData); err != nil {
					continue
				}

				t.Log(tmpData)
			}

			t.Log("\n")

			return nil

		},
		Pointer: func(current *int, boundary int) {
			// change offset
			*current = (*current - 1) * 10
		},
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = pag.Get()

	if err != nil {
		t.Fatalf(err.Error())
	}

}
