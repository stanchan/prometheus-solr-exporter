# Solr Exporter

[![Docker Pulls](https://img.shields.io/docker/pulls/stanchan/prometheus-solr-exporter.svg?maxAge=604800)](https://hub.docker.com/r/stanchan/prometheus-solr-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/stanchan/prometheus-solr-exporter)](https://goreportcard.com/report/github.com/stanchan/prometheus-solr-exporter)

### WARNING: This is a fork of the original noony project... but refactored to support Solr 7+ and up. This was due to the requirement of using this in a large production environment and not being able to support older versions.

Prometheus exporter for various metrics about Solr, written in Go.

### Installation

For pre-built binaries please take a look at the releases.
https://github.com/stanchan/prometheus-solr-exporter

#### Docker

```bash
docker pull stanchan/prometheus-solr-exporter
docker run stanchan/prometheus-solr-exporter --solr.address=http://url-to-solr:port
```

#### Configuration

Below is the command line options summary:

```bash
prometheus-solr-exporter --help
```

| Argument              | Description |
| --------              | ----------- |
| solr.address          | URI on which to scrape Solr. (default "http://localhost:8983") |
| solr.context-path     | Solr webapp context path. (default "/solr") |
| solr.pid-file         | Path to Solr pid file |
| solr.timeout          | Timeout for trying to get stats from Solr. (default 5s) |
| solr.excluded-core    | Regex to exclude core from monitoring|
| web.listen-address    | Address to listen on for web interface and telemetry. (default ":9231")|
| web.telemetry-path    | Path under which to expose metrics. (default "/metrics")|

### Building

Clone the repository and just launch this command
```bash
make build
```

### Testing

[![Build Status](https://travis-ci.org/stanchan/prometheus-solr-exporter.png?branch=master)][travisci]

```bash
make test
```

[travisci]: https://travis-ci.org/stanchan/prometheus-solr-exporter

### Grafana dashboard

See https://grafana.com/dashboards/2551

## License

Apache License 2.0, see [LICENSE](https://github.com/stanchan/prometheus-solr-exporter/blob/master/LICENSE).
