#!/bin/sh
rm memprofile.out
go test -bench -run=^$ -bench ^BenchmarkJSONHandler$ -memprofile=memprofile.out  github.com/keycubasic/shortener/internal/handler
go tool pprof -http=":9090" memprofile.out
