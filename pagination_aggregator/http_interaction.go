package paginationaggregator

import (
	"net/http"
)

type Response struct {
	Status     int
	StatusText string
	Error      error
	Data       string
}

type Request struct {
	Pointer     int
	HttpRequest *http.Request
}

type HttpInteraction struct {
	Response *Response
	Request  *Request
}
