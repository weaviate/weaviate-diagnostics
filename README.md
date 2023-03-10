# Weaviate Diagnostics 🩺

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

Optional recommended depddency is [graphviz](https://graphviz.org/) for the cpu profiling graph.

```sh
brew install graphviz # if on mac
apk add graphviz # if inside weaviate container
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

### To do

- [x] Command line parsing for weaviate url, report file name
- [x] Command line option for prometheus endpoint
- [x] Copy buttons to copy all code sections
- [ ] Use one CDN for js / css assets
- [x] Create list of warning misconfigurations
- [x] Create top level metrics (total memory, total nodes, etc)
- [ ] Memory consumption visualisation
- [x] See if we can use pprof package instead of `go tool pprof` (remove go dependency)
