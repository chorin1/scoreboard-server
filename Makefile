go-get:
	go get ./...

build:
	go build ./...

test:
	go test -cover -race ./...

lint:
	golangci-lint run

run:
	docker-compose up
