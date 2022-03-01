ARG ARCH="amd64"
ARG OS="linux"
FROM golang:1.17 as builder
ENV CGO_ENABLED=0  GOARCH=$ARCH GOOS=$OS

WORKDIR /app
COPY *.go go.mod go.sum /app/
RUN go mod download
COPY *.go /app/
RUN go build -o /app/influxdb_exporter && chmod +x /app/influxdb_exporter

FROM scratch
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"
ARG ARCH="amd64"
ARG OS="linux"
COPY --from=builder /app/influxdb_exporter /bin/influxdb_exporter

EXPOSE      9122
ENTRYPOINT  [ "/bin/influxdb_exporter" ]
