FROM alpine

RUN apk add --no-cache ca-certificates

COPY releases-notifier /releases-notifier

USER nobody

ENTRYPOINT ["/releases-notifier"]
EXPOSE 8080
