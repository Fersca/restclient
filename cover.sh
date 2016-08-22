#!/bin/bash
go test -coverprofile=coverage.out -coverpkg github.com/mercadolibre/go-meli-toolkit/restclient
go tool cover -html=coverage.out
