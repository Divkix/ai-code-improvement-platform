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

openapi-json:
	@echo "Generating OpenAPI JSON from YAML..."
	@command -v yq >/dev/null 2>&1 || { echo >&2 "yq is required but not installed. Install from https://github.com/mikefarah/yq"; exit 1; }
	@yq -o=json ../api/openapi.yaml > ../api/openapi.json
