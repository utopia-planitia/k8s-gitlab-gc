FROM golang:1.10.3-alpine3.7 AS compile
ENV GOOS=linux
COPY vendor /go/src/github.com/utopia-planitia/k8s-gitlab-gc/vendor
RUN go build all
COPY lib /go/src/github.com/utopia-planitia/k8s-gitlab-gc/lib
COPY main.go /go/src/github.com/utopia-planitia/k8s-gitlab-gc/main.go
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/
RUN go install .

FROM alpine:3.8
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
