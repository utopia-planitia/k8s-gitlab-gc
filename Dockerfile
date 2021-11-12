# compile
FROM golang:1.16.5-alpine3.12@sha256:db2475a1dbb2149508e5db31d7d77a75e6600d54be645f37681f03f2762169ba AS compile
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go install .

# package
FROM alpine:3.14.3@sha256:230cdd0ecad7d678b69b033748ac07183a26115ab1050a5d464105eafbe57859
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
