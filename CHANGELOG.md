## 0.11.5 / 2023-12-06

* [SECURITY] Maintenance release, updating dependencies and build with Go 1.21

## 0.11.4 / 2023-06-02

* [SECURITY] Maintenance release, updating dependencies

## 0.11.3 / 2023-03-08

* [Security] Update dependencies ([#131](https://github.com/prometheus/influxdb_exporter/pull/131))
## 0.11.2 / 2023-01-26

This is a maintenance release, updating dependencies.

## 0.11.1 / 2022-12-09

* [BUGFIX] Continue processing UDP after badly formatted packages ([#120](https://github.com/prometheus/influxdb_exporter/pull/120))

## 0.11.0 / 2022-10-25

* [ENHANCEMENT] Improve compatibility of `/ping` endpoint ([#116](https://github.com/prometheus/influxdb_exporter/pull/116))

## 0.10.0 / 2022-07-08

* [FEATURE] Add `/health` endpoint to make more InfluxDB clients happy ([#108](https://github.com/prometheus/influxdb_exporter/pull/108))
* [CHANGE] Update dependencies and build with Go 1.18 ([#109](https://github.com/prometheus/influxdb_exporter/pull/109))

## 0.9.1 / 2022-05-06

This is a maintenance release, built with Go 1.17.9 to address security issues.

## 0.9.0 / 2022-03-08

* [FEATURE] Support ingestion from InfluxDB v2 clients ([#95](https://github.com/prometheus/influxdb_exporter/pull/95))

## 0.8.1 / 2021-11-26

* [ENHANCEMENT] Update Go & dependencies to enable build for windows/arm64 ([#93](https://github.com/prometheus/influxdb_exporter/pull/93))

## 0.8.0 / 2021-01-21

* [ENHANCEMENT] Accept gzip encoding ([#78](https://github.com/prometheus/influxdb_exporter/pull/78))

## 0.7.0 / 2020-12-04

* [ENHANCEMENT] Handle metric names that start with digits ([#77](https://github.com/prometheus/influxdb_exporter/pull/77))

## 0.6.0 / 2020-11-02

* [ENHANCEMENT] Return errors as JSON, in line with InfluxDB ([#74](https://github.com/prometheus/influxdb_exporter/pull/74))

## 0.5.0 / 2020-08-21

* [CHANGE] Move exporter metrics to their own endpoint ([#68](https://github.com/prometheus/influxdb_exporter/pull/68))
* [ENHANCEMENT] Ignore the `__name__` label on incoming metrics ([#69](https://github.com/prometheus/influxdb_exporter/pull/69))

This release improves the experience in mixed Prometheus/InfluxDB environments.
By moving the exporter's own metrics to a separate endpoint, we avoid conflicts with metrics from other services using the Prometheus client library.
In these circumstances, a spurious `__name__` label might appear, which we cannot ingest.
The exporter now ignores it.

## 0.4.2 / 2020-06-12

* [CHANGE] Update all dependencies, including Prometheus client ([#66](https://github.com/prometheus/influxdb_exporter/pull/66))

## 0.4.1 / 2020-05-04

* [ENHANCEMENT] Improve performance by reducing allocations ([#64](https://github.com/prometheus/influxdb_exporter/pull/64))

## 0.4.0 / 2020-02-28

* [FEATURE] Add ping endpoint that some clients expect ([#60](https://github.com/prometheus/influxdb_exporter/pull/60))

## 0.3.0 / 2019-10-04

* [CHANGE] Do not run as root in the Docker container by default ([#40](https://github.com/prometheus/influxdb_exporter/pull/40))
* [CHANGE] Update logging library & flags ([#58](https://github.com/prometheus/influxdb_exporter/pull/58))

## 0.2.0 / 2019-02-28

* [CHANGE] Switch to Kingpin flag library ([#14](https://github.com/prometheus/influxdb_exporter/pull/14))
* [FEATURE] Optionally export samples with timestamp ([#36](https://github.com/prometheus/influxdb_exporter/pull/36))

For consistency with other Prometheus projects, the exporter now expects
POSIX-ish flag semantics. Use single dashes for short options (`-h`) and two
dashes for long options (`--help`).

## 0.1.0 / 2017-07-26

Initial release.
