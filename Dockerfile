FROM alpine:3.4

ADD dist /dist
ADD drone-gcs ./
ENTRYPOINT ["sh"]
