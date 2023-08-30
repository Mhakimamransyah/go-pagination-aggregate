package example

import (
	"encoding/json"
	"net/http"
	"testing"

	paginationaggregator "github.com/Mhakimamransyah/go-pagination-aggregate/pagination_aggregator"
)

type Users struct {
	Id        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
}

type UsersResponse struct {
	Page      int     `json:"page"`
	TotalPage int     `json:"total_pages"`
	PerPage   int     `json:"per_page"`
	Data      []Users `json:"data"`
}

func (obj UsersResponse) GetBoundary() int {
	return obj.TotalPage
}

func TestGetUsers(t *testing.T) {

	pag, err := paginationaggregator.NewPaginationAggregator(&paginationaggregator.PaginationAggregatorConfig{
		URL:               "https://reqres.in/api/users?page=%d&per_page=2",
		JsonPage:          &UsersResponse{},
		Start:             1,
		Client:            &http.Client{},
		DelayBetweenBatch: 3,
		Concurrent:        3,
		ConcurrentBatch: func(batchResult []paginationaggregator.HttpInteraction) error {

			for _, val := range batchResult {

				tmpData := UsersResponse{}

				if err := json.Unmarshal([]byte(val.Response.Data), &tmpData); err != nil {
					continue
				}

				t.Log(tmpData)
			}

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
