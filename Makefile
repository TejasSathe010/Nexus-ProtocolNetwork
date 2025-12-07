APP_NAME=Nexus-ProtocolNetwork

.PHONY: run
run:
	go run ./cmd/gateway

.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/gateway

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy
