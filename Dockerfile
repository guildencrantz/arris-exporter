FROM golang:1.14-alpine AS alpine

WORKDIR /arris-exporter
COPY . .

RUN go build .

WORKDIR /usr/share/zoneinfo
RUN apk --no-cache add tzdata zip ca-certificates && \
    zip -q -r -0 /zoneinfo.zip .

# What is trying to access an external resource so won't work in scratch?
FROM debian:buster
COPY --from=alpine /arris-exporter/arris-exporter /arris-exporter

ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 9100

# NOBODY
#USER 65534

ENTRYPOINT ["/arris-exporter"]
