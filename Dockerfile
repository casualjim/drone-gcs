FROM alpine:3.4

ENV GOPATH=/go
RUN apk update&&\
        apk add go\
        git

RUN mkdir -p /go/src &&\
        cd /go/src &&\
        git clone https://github.com/maplain/drone-gcs &&\
        cd drone-gcs &&\
        go build

ENTRYPOINT ["/go/src/drone-gcs/drone-gcs"]
