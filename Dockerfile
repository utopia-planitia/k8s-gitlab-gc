FROM golang:1.10.3-alpine3.7 AS compile
COPY . /go/src/github.com/utopia-planitia/k8s-gitlab-gc/
WORKDIR /go/src/github.com/utopia-planitia/k8s-gitlab-gc/
RUN CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo .

FROM scratch
COPY --from=compile /go/bin/k8s-gitlab-gc /k8s-gitlab-gc
ENTRYPOINT ["/k8s-gitlab-gc"]
