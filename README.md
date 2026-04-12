# InfluxDB Exporter [![Build Status](https://travis-ci.org/prometheus/influxdb_exporter.svg)][travis]

[![Build Status](https://github.com/prometheus/influxdb_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/prometheus/influxdb_exporter/actions/workflows/ci.yml)
[![Docker Repository on Quay](https://quay.io/repository/prometheus/influxdb-exporter/status)][quay]
[![Docker Pulls](https://img.shields.io/docker/pulls/prom/influxdb-exporter.svg?maxAge=604800)][hub]

An exporter for metrics in the InfluxDB format used since 0.9.0. It collects
metrics in the
[line protocol][line_protocol] via a HTTP API,
transforms them and exposes them for consumption by Prometheus.

This exporter supports float, int and boolean fields. Tags are converted to Prometheus labels.

The exporter also listens on a UDP socket, port 9122 by default, as well as the [v2 HTTP endpoints](#v2-support), where it
exposes InfluxDB metrics using `/metrics` endpoint 
and exposes exporter's self metrics using `/metrics/exporter` endpoint.

## Timestamps

By default metrics exposed without original timestamps like this:

```
http_requests_total{method="post",code="200"} 1027
http_requests_total{method="post",code="400"}    3
```

If you want to add original timestamps to exposed metrics, please use flag `--timestamps` and metrics will looks like:

```
http_requests_total{method="post",code="200"} 1027 1395066363000
http_requests_total{method="post",code="400"}    3 1395066363000
```

When querying, this means that the sample is attributed to the time it was
submitted to the exporter, not the time Prometheus scraped it. However, if the
metric was submitted multiple times in between exporter scrapes, only the last
value and timestamp will be stored.

## Alternatives

If you are sending data to InfluxDB in Graphite or Collectd formats, see the
[graphite_exporter][graphite_exporter]
and [collectd_exporter][collectd_exporter] respectively.

This exporter is useful for exporting metrics from existing collectd setups, as
well as for metrics which are not covered by the core Prometheus exporters such
as the [Node Exporter][node_exporter].

The exporter acts like an InfluxDB server, it does not connect to one. For
metrics concerning the InfluxDB server, use the [metrics endpoint][influxdb_metrics]
built into InfluxDB.

If you are already using Telegraf, it can serve the same purpose as this
exporter with the [`outputs.prometheus_client`][telegraf] plugin.

For more information on integrating between the Prometheus and InfluxDB
ecosystems, see the [influxdata integration page][influx_integration].

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

Note that Telegraf already supports outputting Prometheus metrics over HTTP via
[`outputs.prometheus_client`][telegraf], which avoids having to also run the influxdb_exporter.

## V2 Support
InfluxDB V2 Support is currently in progress. Supported features include:
- Querying for a null result
- Writing data to the exporter - ignores [auth](https://github.com/prometheus/influxdb_exporter/issues/79) and metadata components (org, buckets)

[hub]: https://hub.docker.com/r/prom/influxdb-exporter/
[travis]: https://travis-ci.org/prometheus/influxdb_exporter
[quay]: https://quay.io/repository/prometheus/influxdb-exporter
[line_protocol]: https://docs.influxdata.com/influxdb/v0.10/write_protocols/line/
[graphite_exporter]: https://github.com/prometheus/graphite_exporter
[collectd_exporter]: https://github.com/prometheus/collectd_exporter
[node_exporter]: https://github.com/prometheus/node_exporter
[influxdb_metrics]: https://docs.influxdata.com/influxdb/v1.5/administration/server_monitoring/#influxdb-metrics-http-endpoint
[telegraf]: https://docs.influxdata.com/telegraf/v1.7/plugins/outputs/#prometheus-client-prometheus-client-https-github-com-influxdata-telegraf-tree-release-1-7-plugins-outputs-prometheus-client
[influx_integration]: https://www.influxdata.com/integration/prometheus-monitoring-tool/
