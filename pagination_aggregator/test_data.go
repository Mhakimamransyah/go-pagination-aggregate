package paginationaggregator

import (
	"net/http"
)

type supplyData interface {
	setResponseCollections(data *testData)
}

type animal struct {
	Id     int    `json:"id"`
	Animal string `json:"animal"`
}

// json response with page and size/per page
type jsonTestStructPagePerPage struct {
	StatusCode int      `json:"-"`
	Page       int      `json:"page"`
	TotalPages int      `json:"total_pages"`
	Animals    []animal `json:"data"`
}

func (obj *jsonTestStructPagePerPage) GetBoundary() int {
	return obj.TotalPages
}

type metaTestData struct {
	NumberOfResponse  int
	NumberOfData      int
	NumberOfErrorData int
}

type testData struct {
	Port       int
	Host       string
	Collection []jsonTestStructPagePerPage
	Meta       metaTestData
}

func (obj *testData) addData(newData supplyData) {
	newData.setResponseCollections(obj)
	obj.reFormattingCollectionsPagination()
}

func (obj *testData) reFormattingCollectionsPagination() {

	pages := 1

	for idx := range obj.Collection {
		obj.Collection[idx].Page = pages
		obj.Collection[idx].TotalPages = len(obj.Collection)
		pages++
	}

	meta := metaTestData{
		NumberOfResponse: len(obj.Collection),
		NumberOfData: func() int {
			num := 0
			for _, val := range obj.Collection {
				num += len(val.Animals)
			}
			return num
		}(),
		NumberOfErrorData: func() int {
			num := 0
			for _, val := range obj.Collection {
				if val.StatusCode >= 400 && val.StatusCode <= 599 {
					num++
				}
			}
			return num
		}(),
	}

	obj.Meta = meta
}

func NewTestData(host string, port int, dataSupplier supplyData) *testData {
	var data testData
	data.addData(dataSupplier)
	data.Host = host
	data.Port = port
	return &data
}

// Mocking All Success Response
type SupplySuccessData struct{}

func (obj *SupplySuccessData) setResponseCollections(data *testData) {

	data.Collection = append(data.Collection, []jsonTestStructPagePerPage{
		{
			StatusCode: http.StatusOK,
			Animals: []animal{
				{
					Id:     1,
					Animal: "Anaconda",
				},
				{
					Id:     2,
					Animal: "Bird",
				},
				{
					Id:     3,
					Animal: "Crocodile",
				},
				{
					Id:     4,
					Animal: "Donkey",
				},
				{
					Id:     5,
					Animal: "Eagle",
				},
			},
		},
		{
			StatusCode: http.StatusOK,
			Animals: []animal{
				{
					Id:     6,
					Animal: "Anaconda",
				},
				{
					Id:     7,
					Animal: "Bird",
				},
				{
					Id:     8,
					Animal: "Crocodile",
				},
				{
					Id:     9,
					Animal: "Donkey",
				},
				{
					Id:     10,
					Animal: "Eagle",
				},
			},
		},
		{
			StatusCode: http.StatusOK,
			Animals: []animal{
				{
					Id:     11,
					Animal: "Anaconda",
				},
				{
					Id:     12,
					Animal: "Bird",
				},
				{
					Id:     13,
					Animal: "Crocodile",
				},
				{
					Id:     14,
					Animal: "Donkey",
				},
				{
					Id:     15,
					Animal: "Eagle",
				},
			},
		},
		{
			StatusCode: http.StatusOK,
			Animals: []animal{
				{
					Id:     16,
					Animal: "Anaconda",
				},
				{
					Id:     17,
					Animal: "Bird",
				},
				{
					Id:     18,
					Animal: "Crocodile",
				},
				{
					Id:     19,
					Animal: "Donkey",
				},
				{
					Id:     20,
					Animal: "Eagle",
				},
			},
		},
	}...)
}

// Mocking Error response
type SupplyErrorData struct{}

func (obj *SupplyErrorData) setResponseCollections(data *testData) {

	data.Collection = append(data.Collection, []jsonTestStructPagePerPage{
		{
			StatusCode: http.StatusRequestTimeout,
			Animals:    nil,
		},
	}...)
}
