# compile
FROM golang:1.20.4-alpine3.16@sha256:85c0457d7b6dad9997b5956418bb44fae53c831fe992c7c2964a790467dd8bc0 AS compile
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install .

# package
FROM alpine:3.17.3@sha256:124c7d2707904eea7431fffe91522a01e5a861a624ee31d03372cc1d138a3126
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
