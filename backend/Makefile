TEST_FLAGS ?= -race -count=1 -v

test:
	go test ./... $(TEST_FLAGS)

test-cover:
	go test ./... $(TEST_FLAGS) -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run

generate:
	go generate ./...
