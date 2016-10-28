# InfluxDB Exporter [![Build Status](https://travis-ci.org/prometheus/influxdb_exporter.svg)][travis]

[![CircleCI](https://circleci.com/gh/prometheus/influxdb_exporter/tree/master.svg?style=shield)][circleci]
[![Docker Repository on Quay](https://quay.io/repository/prometheus/influxdb-exporter/status)][quay]
[![Docker Pulls](https://img.shields.io/docker/pulls/prom/influxdb-exporter.svg?maxAge=604800)][hub]

An exporter for metrics in the InfluxDB format used since 0.9.0. It collects
metrics in the
[line protocol](https://docs.influxdata.com/influxdb/v0.10/write_protocols/line/) via a HTTP API,
transforms them and exposes them for consumption by Prometheus.

If you are sending data to InfluxDB in Graphite or Collectd formats, see the
[graphite_exporter](https://github.com/prometheus/graphite_exporter)
and [collectd_exporter](https://github.com/prometheus/collectd_exporter) respectively.

This exporter is useful for exporting metrics from existing collectd setups, as
well as for metrics which are not covered by the core Prometheus exporters such
as the [Node Exporter](https://github.com/prometheus/node_exporter).

This exporter supports float, int and boolean fields. Tags are converted to Prometheus labels.

The exporter also listens on a UDP socket, port 9122 by default.

## Example usage with Telegraf

The influxdb_exporter appears as a normal InfluxDB server. To use with Telegraf
for example, put the following in your `telegraf.conf`:

```
[[outputs.influxdb]]
  urls = ["http://localhost:9122"]
```

Or if you want to use UDP instead:
```
[[outputs.influxdb]]
  urls = ["udp://localhost:9122"]
```

Note that Telegraf already supports outputing Prometheus metrics over HTTP via
`outputs.prometheus_client`, which avoids having to also run the influxdb_exporter.


[circleci]: https://circleci.com/gh/prometheus/influxdb_exporter
[hub]: https://hub.docker.com/r/prom/influxdb-exporter/
[travis]: https://travis-ci.org/prometheus/influxdb_exporter
[quay]: https://quay.io/repository/prometheus/influxdb-exporter
