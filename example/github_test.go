package example

import (
	"net/http"
	"testing"

	paginationaggregator "github.com/Mhakimamransyah/go-pagination-aggregate/pagination_aggregator"
)

func TestGetGithub(t *testing.T) {
	pag, err := paginationaggregator.NewPaginationAggregator(&paginationaggregator.PaginationAggregatorConfig{
		Client: &http.Client{},
		URL:    "https://api.github.com/repositories/1300192/issues?page=%d&per_page=10",
		Headers: paginationaggregator.Header{
			"Accept": "application/vnd.github+json",
		},
		Boundary: 10,
		ConcurrentBatch: func(batchResult []paginationaggregator.HttpInteraction) error {

			for _, val := range batchResult {

				t.Log(val.Response.Data)
			}

			t.Log("\n")

			return nil
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
