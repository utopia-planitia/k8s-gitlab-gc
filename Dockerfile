# compile
FROM golang:1.21.5-alpine AS compile
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install .

# package
FROM alpine:3.19.0@sha256:51b67269f354137895d43f3b3d810bfacd3945438e94dc5ac55fdac340352f48
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
