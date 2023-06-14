# compile
FROM golang:1.20.4-alpine3.16@sha256:85c0457d7b6dad9997b5956418bb44fae53c831fe992c7c2964a790467dd8bc0 AS compile
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install .

# package
FROM alpine:3.18.2@sha256:c7431a7ecef19e2dbab2abd73b052da52de2d3bdce6b32757625a7497f4e93d3
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
