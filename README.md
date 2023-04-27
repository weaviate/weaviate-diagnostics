# Weaviate Diagnostics 🩺

*🚧 This diagnostics tool is still in beta, feedback appreciated at customer-success@weaviate.io. 🚧*

The purpose of this tool is to collect diagnostics from Weaviate clusters and also warn
against any common misconfigurations.

The diagnostics are collected by making a series of requests to a Weaviate instance and
then generating a self-contained html report.

Diagnostics are collected for:

- Weaviate Schema, Meta, Module, and Node config
- pprof CPU profile
- Basic Memory / Disk / CPU info
- Prometheus metrics
- Weaviate specific environment variables

## Dependencies

Optional recommended dependency is [graphviz](https://graphviz.org/) for the cpu profiling graph.

```sh
brew install graphviz # if on mac
apk add graphviz ttf-freefont # if inside weaviate alpine container
```

## Usage

Basic usage for local Weaviate instance:

```sh
./weaviate-diagnostics diagnostics -u "http://localhost:8080" -o weaviate-report.html
```

WCS cluster authenticating using api key

```sh
./weaviate-diagnostics diagnostics -a "$WEAVIATE_API_KEY" -u "https://cluster-name.weaviate.cloud" -o weaviate-report.html
```

Run `-h` for more options:

```sh
./weaviate-diagnostics diagnostics -h
A tool to help diagnose issues with Weaviate

Usage:
  weaviate-diagnostics diagnostics [flags]

Flags:
  -a, --apiKey string       API key authentication
  -h, --help                help for diagnostics
  -m, --metricsUrl string   full URL plus path of the Weaviate metrics endpoint (default "http://localhost:2112/metrics")
  -o, --output string       File to write the report to (default "weaviate-report.html")
  -w, --pass string         Password for OIDC authentication (defaults to prompt)
  -p, --profileUrl string   URL of the Weaviate pprof endpoint (default "http://localhost:6060/debug/pprof/profile?seconds=5")
  -u, --url string          URL of the Weaviate instance (default "http://localhost:8080")
  -n, --user string         Username for OIDC authentication
```
