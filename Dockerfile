# compile
FROM golang:1.22.5-alpine@sha256:523be8675c81c86325fadcf2ff9be68802a716fa94b1e42679c4782e75b076bf AS compile
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install .

# package
FROM alpine:3.20.1@sha256:b89d9c93e9ed3597455c90a0b88a8bbb5cb7188438f70953fede212a0c4394e0
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
