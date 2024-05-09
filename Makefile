# read env vars
include .env
export

.PHONY: lint
lint:
	docker run \
		--rm \
		-w ${PWD} \
		-v ${PWD}:${PWD} \
		docker.io/golangci/golangci-lint:v1.58.1 \
		golangci-lint --timeout 5m0s run ./...

.PHONY: test
test:
	go test $(shell go list ./... | grep -v "/test/e2e")

.PHONY: e2e
e2e:
	go test ./test/e2e
