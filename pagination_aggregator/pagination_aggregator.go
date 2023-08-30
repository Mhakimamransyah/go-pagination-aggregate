package paginationaggregator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	DEFAULT_CONCURRENT = 10
	DEFAULT_TIMEOUT    = 3
	DEFAULT_DELAY      = 2
	DEFAULT_START      = 1
)

type Header map[string]string

type BatchCallback func(batchResult []HttpInteraction) error

type BatchCallbackWithContext func(ctx context.Context, batchResult []HttpInteraction) error

type Pointer func(current *int, boundary int)

type JsonMetaPages interface {
	GetBoundary() int
}

type preProcessingAggregator interface {
	accept(pag *PaginationAggregator) error
}

type PaginationAggregator struct {
	client                     *http.Client
	url                        string
	headers                    Header
	ctx                        context.Context
	start                      int
	boundary                   int
	concurrent                 int
	timeout                    int
	delayBetweenBatch          int
	result                     []HttpInteraction
	concurrentBatch            BatchCallback
	concurrentBatchWithContext BatchCallbackWithContext
	pointer                    Pointer
	jsonPages                  JsonMetaPages
	visitor                    []preProcessingAggregator
}

func (obj *PaginationAggregator) Get() ([]HttpInteraction, error) {

	var err error
	var wg sync.WaitGroup

	if err = obj.runVisitor(); err != nil {
		return nil, err
	}

	channel := make(chan HttpInteraction, obj.concurrent)
	defer close(channel)

	batch := 0
	for pointer := obj.start; pointer <= obj.boundary; pointer++ {

		currentPointer := pointer

		wg.Add(1)

		obj.executePointer(&currentPointer, obj.boundary)

		go obj.fetch(&batch, currentPointer, channel, &wg)

		batch++

		if err = obj.processBatch(&batch, &currentPointer, channel, &wg); err != nil {
			return obj.result, err
		}

		if currentPointer > obj.boundary {
			break
		}

		if obj.ctx != nil && obj.ctx.Err() != nil {
			return nil, obj.ctx.Err()
		}
	}

	return obj.result, nil
}

func (obj *PaginationAggregator) fetch(batch *int, page int, channel chan<- HttpInteraction, wg *sync.WaitGroup) error {

	if page > obj.boundary {

		wg.Done()

		*batch--

		return nil
	}

	var data []byte

	requestCtx, cancel := context.WithTimeout(context.Background(), time.Duration(obj.timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, "GET", fmt.Sprintf(obj.url, page), nil)

	if err != nil {

		channel <- HttpInteraction{
			Request: &Request{
				HttpRequest: req,
				Pointer:     page,
			},
			Response: &Response{
				Status:     http.StatusInternalServerError,
				StatusText: http.StatusText(http.StatusInternalServerError),
				Error:      err,
				Data:       "",
			},
		}

		wg.Done()

		return err
	}

	for key, value := range obj.headers {
		req.Header.Set(key, value)
	}

	resp, err := obj.client.Do(req)

	if err != nil {
		channel <- HttpInteraction{
			Request: &Request{
				HttpRequest: req,
				Pointer:     page,
			},
			Response: &Response{
				Status:     http.StatusInternalServerError,
				StatusText: http.StatusText(http.StatusInternalServerError),
				Error:      err,
				Data:       "",
			},
		}

		wg.Done()

		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {

		var err error

		data, err = io.ReadAll(resp.Body)

		if err == nil && data != nil {
			err = errors.New(string(data))
		} else {
			err = errors.New(http.StatusText(resp.StatusCode))
		}

		channel <- HttpInteraction{
			Request: &Request{
				HttpRequest: req,
				Pointer:     page,
			},
			Response: &Response{
				Status:     resp.StatusCode,
				StatusText: http.StatusText(resp.StatusCode),
				Error:      errors.New(http.StatusText(resp.StatusCode)),
				Data:       "",
			},
		}

		wg.Done()
		return err
	}

	if data, err = io.ReadAll(resp.Body); err != nil {
		channel <- HttpInteraction{
			Request: &Request{
				HttpRequest: req,
				Pointer:     page,
			},
			Response: &Response{
				Status:     http.StatusInternalServerError,
				StatusText: http.StatusText(resp.StatusCode),
				Error:      err,
				Data:       "",
			},
		}

		wg.Done()

		return err
	}

	channel <- HttpInteraction{
		Request: &Request{
			HttpRequest: req,
			Pointer:     page,
		},
		Response: &Response{
			Status:     resp.StatusCode,
			StatusText: resp.Status,
			Data:       string(data),
		},
	}

	wg.Done()

	return nil
}

func (obj *PaginationAggregator) processBatch(batch, currentPointer *int, channel <-chan HttpInteraction, wg *sync.WaitGroup) error {

	if *batch == obj.concurrent || *currentPointer == obj.boundary {

		var tmpBatch []HttpInteraction

		wg.Wait()

		for i := 0; i < *batch; i++ {
			tmpBatch = append(tmpBatch, <-channel)
		}

		obj.result = append(obj.result, tmpBatch...)

		if err := obj.executeCallback(tmpBatch); err != nil {
			return err
		}

		if *currentPointer != obj.boundary {
			time.Sleep(time.Duration(obj.delayBetweenBatch) * time.Second)
		}

		*batch = 0
	}

	return nil

}

func (obj *PaginationAggregator) executePointer(currentPointer *int, boundary int) {

	if obj.pointer != nil {
		obj.pointer(currentPointer, obj.boundary)
	}
}

func (obj *PaginationAggregator) fillDefault() *PaginationAggregator {

	if obj.concurrent == 0 {
		obj.concurrent = DEFAULT_CONCURRENT
	}

	if obj.start == 0 {
		obj.start = DEFAULT_START
	}

	if obj.timeout == 0 {
		obj.timeout = DEFAULT_TIMEOUT
	}

	if obj.delayBetweenBatch == 0 {
		obj.delayBetweenBatch = DEFAULT_DELAY
	}

	return obj
}

func (obj *PaginationAggregator) runVisitor() error {

	for _, visit := range obj.visitor {
		if err := visit.accept(obj); err != nil {
			return err
		}
	}

	return nil
}

func (obj *PaginationAggregator) executeCallback(tmpBatch []HttpInteraction) error {

	var callbackErr error

	if obj.concurrentBatch != nil {
		callbackErr = obj.concurrentBatch(tmpBatch)
	}

	if obj.concurrentBatchWithContext != nil {
		callbackErr = obj.concurrentBatchWithContext(obj.ctx, tmpBatch)
	}

	return callbackErr
}
