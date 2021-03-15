.PHONY: run
run:
	go run cmd/challenge/server.go

.PHONY: lint
lint:
	golangci-lint run ./... -v

.PHONY: test
test:
	go run cmd/challenge/server.go &
	go test -race -v ./cmd/challenge/...
