FROM docker.io/library/alpine:3.22 as runtime

RUN \
  apk add --update --no-cache \
    bash \
    curl \
    ca-certificates \
    tzdata

ENTRYPOINT ["provider-virtuozzo"]
CMD ["operator"]
COPY provider-virtuozzo /usr/bin/

USER 65536:0
