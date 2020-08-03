FROM golang:1.14-buster AS buster

WORKDIR /arris-exporter
COPY . .

RUN go build .

# What is trying to access an external resource so won't work in scratch?
FROM debian:buster
COPY --from=buster /arris-exporter/arris-exporter /arris-exporter

EXPOSE 9100

# NOBODY
#USER 65534

ENTRYPOINT ["/arris-exporter"]
