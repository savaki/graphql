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

```golang
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

## Refs

* [graphql working draft](http://facebook.github.io/graphql/) - 2015.07.02

