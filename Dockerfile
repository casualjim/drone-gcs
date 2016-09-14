FROM alpine:3.4

ADD drone-gcs /bin/
ENTRYPOINT ["/bin/drone-gcs"]
