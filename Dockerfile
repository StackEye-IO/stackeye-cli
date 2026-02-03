FROM alpine:3.21

RUN apk add --no-cache ca-certificates && \
    adduser -D -h /home/stackeye stackeye

COPY stackeye /usr/bin/stackeye

USER stackeye

ENTRYPOINT ["stackeye"]
