package paginationaggregator

import (
	"context"
	"errors"
	"net/http"
)

type PaginationAggregatorConfig struct {
	// http client
	Client *http.Client

	// API url with integer placeholder (%d)
	URL string

	// Requests header
	Headers Header

	// Start page/offset
	Start int

	// End page/offset
	Boundary int

	// Number of concurrent requests for every batch
	Concurrent int

	// Requests timeout in seconds
	Timeout int

	// Delay time on every batch requests in seconds
	DelayBetweenBatch int

	//	Override this function to manage behaviour for every batch requests
	ConcurrentBatch BatchCallback

	//	Override this function to manage behaviour for every batch requests with injected context
	ConcurrentBatchWithContext BatchCallbackWithContext

	// Override this function to manage behaviour every pointer iteration
	Pointer Pointer

	// Struct which bind single json response to retrieve pagination boundary
	JsonPage JsonMetaPages

	visitor []preProcessingAggregator
}

func NewPaginationAggregator(config *PaginationAggregatorConfig) (*PaginationAggregator, error) {

	if err := config.tidyUpConfigurations(); err != nil {
		return nil, err
	}

	pag := PaginationAggregator{
		client:            config.Client,
		url:               config.URL,
		start:             config.Start,
		boundary:          config.Boundary,
		headers:           config.Headers,
		delayBetweenBatch: config.DelayBetweenBatch,
		concurrent:        config.Concurrent,
		timeout:           config.Timeout,
		concurrentBatch:   config.ConcurrentBatch,
		pointer:           config.Pointer,
		jsonPages:         config.JsonPage,
		visitor:           config.visitor,
	}

	return pag.fillDefault(), nil
}

func NewPaginationAggregatorWithContext(ctx context.Context, config *PaginationAggregatorConfig) (*PaginationAggregator, error) {

	if err := config.tidyUpConfigurations(); err != nil {
		return nil, err
	}

	pag := PaginationAggregator{
		client:                     config.Client,
		url:                        config.URL,
		start:                      config.Start,
		boundary:                   config.Boundary,
		headers:                    config.Headers,
		delayBetweenBatch:          config.DelayBetweenBatch,
		concurrent:                 config.Concurrent,
		pointer:                    config.Pointer,
		timeout:                    config.Timeout,
		concurrentBatchWithContext: config.ConcurrentBatchWithContext,
		ctx:                        ctx,
		jsonPages:                  config.JsonPage,
		visitor:                    config.visitor,
	}

	return pag.fillDefault(), nil
}

func (obj *PaginationAggregatorConfig) tidyUpConfigurations() error {

	if obj.Boundary == 0 {
		obj.visitor = append(obj.visitor, newBoundaryAssertion())
	}

	if obj.URL == "" {
		return errors.New("No Http URL Found")
	}

	if obj.Client == nil {
		return errors.New("No Http Client Found")
	}

	return nil
}
