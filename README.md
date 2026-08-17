# dyndns

Dyndns Library

## Install

```
go get go-mw.de/pkg/dyndns
```

```go
import "go-mw.de/pkg/dyndns"
```

`go-mw.de/pkg/dyndns` is a vanity import path. The meta tags that point the Go
toolchain at this repository are served by the
[appengine](https://github.com/go-mw-de/appengine) project, so the import path stays
stable even if the code moves to another host. The previous paths
`github.com/go-mw-de/dyndns` and `gitlab.com/go-mw-de/dyndns` are no longer
importable; update your imports and run `go mod tidy`.
