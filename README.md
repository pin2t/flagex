# flagex
[![GoDoc](https://godoc.org/github.com/pin2t/flagex?status.svg)](https://godoc.org/github.com/pin2t/flagex)


Extended flag parsing for Go. Reads command-line flags from configuration files, streams, and URLs on top of the standard `flag` package.

## Usage

```go
package main

import "flag"
import "github.com/pin2t/flagex"

var user = flag.String("user", "", "user name")
var port = flag.Int("port", 8080, "port number")

func main() {
    flag.Parse()
    flagex.ParseFile("app.conf")
    flagex.ParseURL("http://config-server/app.conf")
}
```

## Format

Each line is `name=value`. Blank lines and lines starting with `#` are ignored. Flags already set on the command line take precedence.

## Functions

- **ParseFile(path)** — reads configuration from a file.
- **ParseStream(r)** — reads configuration from an `io.Reader`.
- **ParseURL(url)** — fetches configuration via HTTP GET.
- **ParseEnv(prefix)** — fetches configuration from system environment variables.
