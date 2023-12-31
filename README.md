# go-pagination-aggregate
[![Go Reference](https://pkg.go.dev/badge/github.com/Mhakimamransyah/go-pagination-aggregate.svg)](https://pkg.go.dev/github.com/Mhakimamransyah/go-pagination-aggregate) 
[![codecov](https://codecov.io/gh/Mhakimamransyah/go-pagination-aggregate/graph/badge.svg?token=SJHKHOYMDL)](https://codecov.io/gh/Mhakimamransyah/go-pagination-aggregate) ![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/Mhakimamransyah/go-pagination-aggregate?logo=go)  [![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=mhakimamransyah_mhakim_7agu45&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=mhakimamransyah_mhakim_7agu45)[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=mhakimamransyah_mhakim_7agu45&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=mhakimamransyah_mhakim_7agu45) [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=mhakimamransyah_mhakim_7agu45&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=mhakimamransyah_mhakim_7agu45)

A Library allows you to merge all api json responses which implement pagination into single data structures. 
there are many api pagination mechanisms that let us interact with it. So far this library **only supports page-sized and offset-limit pagination** 
which use integer values to iterate and shift data pointer.

## Contents
- [Usage](https://github.com/Mhakimamransyah/go-pagination-aggregate#usage)
- [Set up and configure instance](https://github.com/Mhakimamransyah/go-pagination-aggregate#setup-and-configure-instance)
- [Configure asynchronous requests](https://github.com/Mhakimamransyah/go-pagination-aggregate#configure-asynchronous-requests)
- [Pagination with limit and offset params](https://github.com/Mhakimamransyah/go-pagination-aggregate#pagination-with-limit-and-offset-params)
- [Example](https://github.com/Mhakimamransyah/go-pagination-aggregate#example)
  
## Usage
You need to defined full url path in strings with integer placeholder which will use as iterator request, 
example :
```
https://your.api.com?page=%d&per_page=100
```
In case you dont want to specify last pages value using your own, you need to implement this interface that represent single json response 
```
type JsonMetaPages interface {
    GetBoundary() int
}
```
or you can just specify last pages of your paginate api with ```boundary``` configurations.
All request will work asynchronously for every batch with some configurations need on it.
### Setup and Configure Instance
Let says you need to aggregate all api paginated response from this url
```
https://reqres.in/api/users?page=1&per_page=5
```
Response
```
{
  "page": 1,
  "per_page": 5,
  "total": 12,
  "total_pages": 3,
  "data": [
    {
      "id": 1,
      "email": "george.bluth@reqres.in",
      "first_name": "George",
      "last_name": "Bluth",
      "avatar": "https://reqres.in/img/faces/1-image.jpg"
    },
    {
      "id": 2,
      "email": "janet.weaver@reqres.in",
      "first_name": "Janet",
      "last_name": "Weaver",
      "avatar": "https://reqres.in/img/faces/2-image.jpg"
    },
    {
      "id": 3,
      "email": "emma.wong@reqres.in",
      "first_name": "Emma",
      "last_name": "Wong",
      "avatar": "https://reqres.in/img/faces/3-image.jpg"
    },
    {
      "id": 4,
      "email": "eve.holt@reqres.in",
      "first_name": "Eve",
      "last_name": "Holt",
      "avatar": "https://reqres.in/img/faces/4-image.jpg"
    },
    {
      "id": 5,
      "email": "charles.morris@reqres.in",
      "first_name": "Charles",
      "last_name": "Morris",
      "avatar": "https://reqres.in/img/faces/5-image.jpg"
    }
  ],
  "support": {
    "url": "https://reqres.in/#support-heading",
    "text": "To keep ReqRes free, contributions towards server costs are appreciated!"
  }
}
```

create an instance and get the results
```
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

pag, err := NewPaginationAggregator(&PaginationAggregatorConfig{
 	Client: &http.Client{},
	URL: "https://reqres.in/api/users?page=%d&per_page=5",
	JsonPage: &UsersResponse{},
})

if err != nil {
   fmt.Println(err.Error())
   return
}

users, err := pag.Get()

if err != nil {
   fmt.Println(err.Error())
   return
}

// loop thorugh all responses
for _, val := range users {
    tmpData := UsersResponse{}
    if err := json.Unmarshal([]byte(val.Response.Data), &tmpData); err != nil {
        continue
    }
    fmt.Println(tmpData)
}
```
### Configure asynchronous requests
You can define number of concurrent requests that will be made using ```Concurrent``` configuration
```
pag, err := NewPaginationAggregator(&PaginationAggregatorConfig{
	Client: &http.Client{},
	URL: "https://your.pagination.com?page=%d&per_page=5",
	JsonPage: &UsersResponse{},
	Concurrent: 2,
})
```
if you have to make 100 requests and you set ```Concurrent = 50``` then it will make 2 batch asynchronous requests. 
You can also manipulate data for every batch with overriding  ```ConcurrentBatch``` function like this.
```
pag, err := NewPaginationAggregator(&PaginationAggregatorConfig{
	Client: &http.Client{},
	URL: "https://your.pagination.com?page=%d&per_page=5",
	JsonPage: &UsersResponse{},
    	Concurrent: 2,
    	ConcurrentBatch: func(batchResult []paginationaggregator.HttpInteraction) error {

       		for _, val := range batchResult {
			 tmpData := UsersResponse{}
			 if err := json.Unmarshal([]byte(val.Response.Data), &tmpData); err != nil {
				 continue
			 }
         		// INSERT tmpData TO DB ...
		}

       // if return error then it will stop to processing all next batch requests and returning an error, otherwise it will continue processing batch requests
       return nil

    },
})
```
### Pagination with limit and offset params
You can manipulate integer iterator value using override ```Pointer``` function like this
```
pag, err := NewPaginationAggregator(&PaginationAggregatorConfig{
	Client: &http.Client{},
	URL: "https://your.pagination.com?page=%d&per_page=5",
	JsonPage: &UsersResponse{},
    	Pointer : func(current *int, boundary int) {
		*current = *current + 2
	},
})
```
it will change requests page which add 2 on every pages so when pages 1 it will requests pages 3. In case you need to consume paginated api response with limit-offset params you can use
```
pag, err := NewPaginationAggregator(&PaginationAggregatorConfig{
	Client: &http.Client{},
	URL: "https://your.pagination.com?offset=%d&limit=10",
	JsonPage: &UsersResponse{},
    	Pointer : func(current *int, boundary int) {
		// change offset
		*current = (*current - 1) * 10
	},
})
```
it will shift 10 number offset value while keeping limit size

### Example 
- [Simple usage](https://codefile.io/f/GhnGDEP9Y6)
- [Insert response to database for each batch](https://codefile.io/f/QuGgwhWSO2)




