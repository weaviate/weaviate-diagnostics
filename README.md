# Weaviate Diagnostics

The purpose of this tool is to collect diagnostics from Weaviate instances
running in customer environments.

## Dependencies

Go needs to be installed for CPU profile. Ideally this tool is run from
the Weaviate instance 

```sh
brew install golang # if on mac
apk add golang # if inside weaviate container
```

## Usage

```sh
go run main.go > weaviate-report.html
```

### To do

- [ ] All inputs are optional and fail gracefully
- [ ] Command line parsing for weaviate url, report file name
- [ ] Command line option for prometheus endpoint
- [ ] Copy buttons to copy all code sections
- [ ] Default to report-TIMESTAMP.html for file name
- [ ] Use one CDN for js / css assets
- [ ] Parse prometheus metrics in Golang
- [ ] Create list of warning misconfigurations
- [ ] Create top level metrics (total memory, total nodes, etc)
- [ ] Memory consumption visualisation
- [ ] See if we can use pprof package instead of `go tool pprof` (remove go dependency)