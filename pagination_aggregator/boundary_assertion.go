package paginationaggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BoundaryAssertion struct {
}

func (obj *BoundaryAssertion) accept(pag *PaginationAggregator) error {

	var client = &http.Client{}
	var url = fmt.Sprintf(pag.url, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pag.timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return err
	}

	for key, value := range pag.headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&pag.jsonPages); err != nil {
		return err
	}

	pag.boundary = pag.jsonPages.GetBoundary()

	return nil
}

func newBoundaryAssertion() *BoundaryAssertion {
	return &BoundaryAssertion{}
}
