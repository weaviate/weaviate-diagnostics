# Weaviate Diagnostics

The purpose of this tool is to collect diagnostics from Weaviate instances
running in customer environments.

## Dependencies

Optional recommended depdency is [graphviz](https://graphviz.org/) for the cpu profiling graph.

```sh
brew install graphviz # if on mac
apk add graphviz # if inside weaviate container
```

## Usage

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

- [ ] All inputs are optional and fail gracefully
- [ ] Command line parsing for weaviate url, report file name
- [ ] Command line option for prometheus endpoint
- [ ] Copy buttons to copy all code sections
- [ ] Default to report-TIMESTAMP.html for file name
- [ ] Use one CDN for js / css assets
- [ ] Parse prometheus metrics in Golang
- [x] Create list of warning misconfigurations
- [ ] Create top level metrics (total memory, total nodes, etc)
- [ ] Memory consumption visualisation
- [x] See if we can use pprof package instead of `go tool pprof` (remove go dependency)
