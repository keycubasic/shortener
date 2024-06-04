#!/bin/sh
rm coverage.out
rm tmp/short-url-db.json
go test -v -tags=integration -coverpkg=./... -coverprofile=coverage.out -covermode=count ./...
go tool cover -func coverage.out | grep total | awk '{print $3}'
