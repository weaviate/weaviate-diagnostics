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

Authenticated cluster with OpenID (will prompt to ask for credentials):

```sh
./weaviate-diagnostics diagnostics -u "https://cluster-name.weaviate.cloud" -o weaviate-report.html -a
```

Run `-h` for more options:

```sh
./weaviate-diagnostics -h
A tool to help diagnose issues with Weaviate

Usage:
  weaviate-diagnostics [flags]
  weaviate-diagnostics [command]

Available Commands:
  diagnostics Run Weaviate Diagnostics
  help        Help about any command
  profile     Generate a CPU profile

Flags:
  -h, --help   help for weaviate-diagnostics

Use "weaviate-diagnostics [command] --help" for more information about a command.
```
