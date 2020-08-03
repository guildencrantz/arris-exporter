FROM golang:1.14-buster AS builder

COPY . /arris-exporter
WORKDIR /arris-exporter

RUN go build .

FROM scratch
COPY --from=builder /arris-exporter/arris-exporter /arris-exporter

EXPOSE 9100

# NOBODY
USER 65534

ENTRYPOINT ["/arris-exporter"]
