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

### File example

```conf
# app.conf
user=configuser
url=http://localhost:8332
listen=:9999
count=42
verbose=true
ratio=0.75
```

### Environment variables example

With prefix `APP_`, environment variables map to flag names by stripping the prefix
and converting to lowercase. For example, `APP_URL` sets the `url` flag.

```sh
export APP_USER=envuser
export APP_URL=http://env:9000
export APP_LISTEN=:8888
export APP_COUNT=99
export APP_VERBOSE=true
export APP_RATIO=1.5
```

## Functions

- **ParseFile(path)** — reads configuration from a file.
- **ParseStream(r)** — reads configuration from an `io.Reader`.
- **ParseURL(url)** — fetches configuration via HTTP GET.
- **ParseEnv(prefix)** — fetches configuration from system environment variables.

## FlagSet API

The `FlagSet` function wraps a `flag.FlagSet` so you can parse configuration into
a non-default FlagSet (e.g., one created with `flag.NewFlagSet`).

```go
package main

import "flag"
import "github.com/pin2t/flagex"

func main() {
    fs := flag.NewFlagSet("myapp", flag.ExitOnError)
    user := fs.String("user", "", "user name")
    port := fs.Int("port", 8080, "port number")

    flagex.FlagSet(fs).ParseFile("app.conf")
    flagex.FlagSet(fs).ParseEnv("APP_")

    // Flags from file and env are set; command-line values take precedence.
    fmt.Println(*user, *port)
}
```

All four configuration sources are available as methods on the wrapper:

- **FlagSet(fs).ParseFile(path)** — reads configuration from a file.
- **FlagSet(fs).ParseStream(r)** — reads configuration from an `io.Reader`.
- **FlagSet(fs).ParseURL(url)** — fetches configuration via HTTP GET.
- **FlagSet(fs).ParseEnv(prefix)** — reads configuration from environment variables.

The package-level `ParseFile`, `ParseStream`, `ParseURL`, and `ParseEnv` functions
are equivalent to calling the methods on `FlagSet(flag.CommandLine)`.
