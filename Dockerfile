# compile
FROM golang:1.22.3-alpine@sha256:f1fe698725f6ed14eb944dc587591f134632ed47fc0732ec27c7642adbe90618 AS compile
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install .

# package
FROM alpine:3.20.0@sha256:77726ef6b57ddf65bb551896826ec38bc3e53f75cdde31354fbffb4f25238ebd
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
