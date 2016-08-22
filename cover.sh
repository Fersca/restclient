#!/bin/bash
go test -coverprofile=coverage.out -coverpkg github.com/Fersca/restclient
go tool cover -html=coverage.out
