set shell := ["bash", "-cu"]

default:
  @just --list

build:
  go build -o vja .

dev:
  go run .

test:
  go test ./...

fmt:
  go fmt ./...

tidy:
  go mod tidy

check: fmt test
