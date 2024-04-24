run:
	go run ./cmd/main.go --config=./local.yaml

build:
	go build ./cmd/main.go

start:
	./main --config=./local.yaml

lint:
	golangci-lint run