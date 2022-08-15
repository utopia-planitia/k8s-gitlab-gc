# read env vars
include .env
export

.PHONY: lint
lint:
	docker run \
		-ti \
		--rm \
		-w ${PWD} \
		-v ${PWD}:${PWD} \
		--env GOFLAGS=-buildvcs=false \
		golangci/golangci-lint:v1.48.0-alpine \
		golangci-lint --timeout=540s run ./...
.PHONY: test
test:
	go test ./...
