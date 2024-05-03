#!/bin/sh
swag fmt
swag init --generalInfo internal/handler/handler.go --output cmd/docs --parseDependency --parseInternal
# go run cmd/shortener/main.go
# start http://localhost:8080/swagger/index.html
