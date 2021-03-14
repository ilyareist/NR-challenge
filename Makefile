.PHONY: run
run:
	go run cmd/challenge/server.go

.PHONY: lint
lint:
	golangci-lint run ./... -v

.PHONY: test
test:
	go test -race -v -covermode=atomic -coverpkg=./internal/...,./pkg/... -coverprofile=coverage.out ./tests/... ./pkg/...
