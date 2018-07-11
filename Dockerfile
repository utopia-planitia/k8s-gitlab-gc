FROM golang:1.10.3-alpine3.7 AS compile
COPY . /go/src/github.com/utopia-planitia/kubernetes-gitlab-garbage-collector/
WORKDIR /go/src/github.com/utopia-planitia/kubernetes-gitlab-garbage-collector/
RUN CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo ./cmd/garbage-collector

FROM scratch
COPY --from=compile /go/bin/garbage-collector /garbage-collector
ENTRYPOINT ["/garbage-collector"]
