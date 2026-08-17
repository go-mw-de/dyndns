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

## Releasing

Push a `vMAJOR.MINOR.PATCH` tag. A workflow then asks `proxy.golang.org` for that
version, so it is cached while go-mw.de is known to be healthy — a module version
in the proxy is served without ever contacting go-mw.de again, which is what keeps
an outage of the vanity handler from blocking builds. If the job fails, go-mw.de
is not serving the meta tags; see `docs/vanity-imports.md` in the
[appengine](https://github.com/go-mw-de/appengine) repository for the recovery
procedure.
