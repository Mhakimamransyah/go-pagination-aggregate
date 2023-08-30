package paginationaggregator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

var testTables *testData

func TestGetWithSuccessResponse(t *testing.T) {

	var concurrentRequest = 2
	var acceptType = "application/json"

	pag, err := NewPaginationAggregator(&PaginationAggregatorConfig{
		Client:     &http.Client{},
		JsonPage:   &jsonTestStructPagePerPage{},
		Concurrent: concurrentRequest,
		Headers: Header{
			"Accept": acceptType,
		},
		URL: testTables.Host + ":" + strconv.Itoa(testTables.Port) + "/data?page=%d",
		ConcurrentBatch: func(batchResult []HttpInteraction) error {

			if len(batchResult) != concurrentRequest {
				t.Fatalf("Data collected on each batch not same as num of concurrent request, expected %d actual %d", concurrentRequest, len(batchResult))
			}

			return nil
		},
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	result, err := pag.Get()

	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(result) != testTables.Meta.NumberOfResponse {
		t.Errorf("Data collected different from original data, expected %d actual %d", len(result), testTables.Meta.NumberOfResponse)
	}

	for _, val := range result {
		if val.Request.HttpRequest.Header["Accept"][0] != acceptType {
			t.Errorf("Get different requests header, expected %s actual %s", acceptType, val.Request.HttpRequest.Header["accept"][0])
		}
	}
}

func TestGetWithErrorResponse(t *testing.T) {

	var errorRequest HttpInteraction

	testTables.addData(&SupplyErrorData{})

	pag, err := NewPaginationAggregatorWithContext(context.Background(), &PaginationAggregatorConfig{
		Client:   &http.Client{},
		JsonPage: &jsonTestStructPagePerPage{},
		URL:      testTables.Host + ":" + strconv.Itoa(testTables.Port) + "/data?page=%d",
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	response, err := pag.Get()

	if err != nil {
		t.Fatalf(err.Error())
	}

	numOfError := 0
	for _, val := range response {
		if val.Response.Error != nil {
			numOfError++
			errorRequest = val
		}
	}
	if numOfError != testTables.Meta.NumberOfErrorData {
		t.Errorf("Number of error response not match, expected %d actual %d", numOfError, testTables.Meta.NumberOfErrorData)
	}
	if errorRequest.Response.Status != http.StatusRequestTimeout {
		t.Errorf("Http Error response not match, expected %d but actual %d", http.StatusRequestTimeout, errorRequest.Response.Status)
	}
}

func TestErrorConfigurations(t *testing.T) {
	t.Run("Config error with no context", func(t *testing.T) {
		_, err := NewPaginationAggregator(&PaginationAggregatorConfig{})
		if err == nil {
			t.Errorf("Error must not be null")
		}
	})
	t.Run("Config error with context", func(t *testing.T) {
		_, err := NewPaginationAggregatorWithContext(context.Background(), &PaginationAggregatorConfig{})
		if err == nil {
			t.Errorf("Error must not be null")
		}
	})

	t.Run("Config error client not set", func(t *testing.T) {
		_, err := NewPaginationAggregatorWithContext(context.Background(), &PaginationAggregatorConfig{
			URL: testTables.Host + ":" + strconv.Itoa(testTables.Port) + "/data?page=%d",
		})
		if err == nil {
			t.Errorf("Error must not be null")
		}
	})
}

func TestStopProcessRequestInCallbackFunc(t *testing.T) {

	customErr := errors.New("Stop processing next Batch")

	t.Run("calback stop with no context", func(t *testing.T) {
		pag, err := NewPaginationAggregator(&PaginationAggregatorConfig{
			Client:     &http.Client{},
			JsonPage:   &jsonTestStructPagePerPage{},
			URL:        testTables.Host + ":" + strconv.Itoa(testTables.Port) + "/data?page=%d",
			Concurrent: 2,
			ConcurrentBatch: func(batchResult []HttpInteraction) error {
				return customErr
			},
		})

		if err != nil {
			t.Fatalf(err.Error())
		}

		res, err := pag.Get()

		if err == nil {
			t.Errorf("Error must not null, expected %s actual %s", customErr.Error(), err.Error())
		}

		if len(res) != 2 {
			t.Errorf("Response collected not match, expected %d actual %d", 2, len(res))
		}
	})

	t.Run("calback stop with context", func(t *testing.T) {
		pag, err := NewPaginationAggregatorWithContext(context.Background(), &PaginationAggregatorConfig{
			Client:     &http.Client{},
			JsonPage:   &jsonTestStructPagePerPage{},
			URL:        testTables.Host + ":" + strconv.Itoa(testTables.Port) + "/data?page=%d",
			Concurrent: 2,
			ConcurrentBatchWithContext: func(ctx context.Context, batchResult []HttpInteraction) error {
				return customErr
			},
		})

		if err != nil {
			t.Fatalf(err.Error())
		}

		res, err := pag.Get()

		if err == nil {
			t.Errorf("Error must not null, expected %s actual %s", customErr.Error(), err.Error())
		}

		if len(res) != 2 {
			t.Errorf("Response collected not match, expected %d actual %d", 2, len(res))
		}
	})
}

func TestStopProcessRequestOverlapPointer(t *testing.T) {

	pag, err := NewPaginationAggregator(&PaginationAggregatorConfig{
		Client:   &http.Client{},
		JsonPage: &jsonTestStructPagePerPage{},
		URL:      testTables.Host + ":" + strconv.Itoa(testTables.Port) + "/data?page=%d",
		Pointer: func(current *int, boundary int) {
			*current = *current + 3
		},
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	res, err := pag.Get()

	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(res) != 2 {
		t.Errorf("Collection of response must only 2 actual %d", len(res))
	}

	requestedPage := map[int]int{
		4: 4,
		5: 5,
	}

	statusCode := map[int]int{
		http.StatusOK:             http.StatusOK,
		http.StatusRequestTimeout: http.StatusRequestTimeout,
	}

	for _, val := range res {
		if _, ok := requestedPage[val.Request.Pointer]; !ok {
			t.Fatalf("Requested pointer not match")
		}

		if _, ok := statusCode[val.Response.Status]; !ok {
			t.Fatalf("Http Response code not match")
		}
	}
}

func TestGetWithtTimeoutContext(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	pag, err := NewPaginationAggregatorWithContext(ctx, &PaginationAggregatorConfig{
		Client:     &http.Client{},
		JsonPage:   &jsonTestStructPagePerPage{},
		URL:        testTables.Host + ":" + strconv.Itoa(testTables.Port) + "/data?page=%d",
		Concurrent: 2,
		Pointer: func(current *int, boundary int) {
			// add sleep in pointer function
			time.Sleep(2 * time.Second)
		},
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = pag.Get()

	if err == nil {
		t.Errorf("Error must not be null")
	}

}

func TestMain(t *testing.M) {

	port := 1234
	host := "http://localhost"

	testTables = NewTestData(host, port, &SupplySuccessData{})

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {

		var response []byte

		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "invalid request")
			return
		}

		pageString := query.Get("page")
		page, err := strconv.Atoi(pageString)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		for _, val := range testTables.Collection {
			if val.Page == page {
				w.WriteHeader(val.StatusCode)
				response, _ = json.Marshal(val)
				break
			}
		}

		fmt.Fprintf(w, string(response))

		return

	})

	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	t.Run()
}
