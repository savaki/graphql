# gographql

[![GoDoc](https://godoc.org/github.com/savaki/graphql?status.svg)](https://godoc.org/github.com/savaki/graphql)

GraphQL implementation in go based on the working draft.

## Status

* The code works, but is far from production grade and does not implement the entire working draft.
* Exploring models for an executor, expect this to change

## Overview

This is a Go implementation of GraphQL.  The intent is to create a high quality library suitable for production deployments.

## Hello World

The famous "hello world" in graphql:

```go
package main

import (
	"os"

	"github.com/savaki/graphql"
	"github.com/savaki/graphql/provider/mapq"
)

func main() {
	model := map[string]interface{}{"hello": "world"}
	store := mapq.New(model)
	graphql.New(store).Handle(`{hello}`, os.Stdout)
	// prints {"hello":"world"}
}
```

## Store

To implement a graphql service, one needs to implement the ```graphql.Store``` interface.  For convenience and as examples, a number of default Store implementations are provided:

* ```github.com/savaki/graphql/provider/mapq``` - access static  ```map[string]interface{}```
* ```github.com/savaki/graphql/provider/jsonq``` - provides a rest gateway

## Rest Call

Here's an example using the ```jsonq``` provider to access a generic rest service.

```go
package main

import (
	"os"

	"github.com/savaki/graphql/provider/restq"
	"github.com/savaki/graphql"
)

func main() {
	query := `query city: GET("http://api.openweathermap.org/data/2.5/weather?lat=35&lon=139") {
		name
		weather: main {
			temperature: temp
		}
	}`

	store := restq.New()
	graphql.New(store).Handle(query, os.Stdout)
}
```

## Refs

* [graphql working draft](http://facebook.github.io/graphql/) - 2015.07.02

