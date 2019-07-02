ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/influxdb_exporter /bin/influxdb_exporter

USER        nobody
EXPOSE      9122
ENTRYPOINT  [ "/bin/influxdb_exporter" ]
