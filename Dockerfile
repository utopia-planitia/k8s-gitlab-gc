# compile
FROM golang:1.24.5-alpine@sha256:ddf52008bce1be455fe2b22d780b6693259aaf97b16383b6372f4b22dd33ad66 AS compile
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install .

# package
FROM alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
