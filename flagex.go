// Package flagex extends the standard flag package with functions to parse
// command-line flags from configuration files, streams, and URLs.
//
// Each line in the configuration should be in name=value format. Blank lines
// and lines starting with # are ignored. Flags already set via the command
// line take precedence and are not overwritten.
package flagex

import "flag"
import "fmt"
import "io"
import "net/http"
import "os"
import "strings"

var parseConfig = func(source string, data string) error {
    var set = make(map[string]bool)
    flag.Visit(func(f *flag.Flag) { set[f.Name] = true })
    for i, line := range strings.Split(data, "\n") {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") { continue }
        var name, value, found = strings.Cut(line, "=")
        if !found { return fmt.Errorf("%s:%d: expected name=value, got %q", source, i+1, line) }
        name = strings.TrimSpace(name)
        value = strings.TrimSpace(value)
        if flag.Lookup(name) == nil { return fmt.Errorf("%s:%d: unknown option %q", source, i+1, name) }
        if set[name] { continue }
        if err := flag.Set(name, value); err != nil { return fmt.Errorf("%s:%d: %v", source, i+1, err) }
    }
    return nil
}

// ParseFile reads a configuration file at path and sets any flags defined in the
// standard flag package that have not already been set on the command line.
func ParseFile(path string) error {
    var data, err = os.ReadFile(path)
    if err != nil { return err }
    return parseConfig(path, string(data))
}

// ParseStream reads configuration from r and sets flags in the standard flag
// package. Flags already set via the command line are not overwritten.
func ParseStream(r io.Reader) error {
    var data, err = io.ReadAll(r)
    if err != nil { return err }
    return parseConfig("stream", string(data))
}

// ParseURL fetches configuration from url via HTTP GET and sets flags.
// A non-200 status code returns an error. Flags already set on the command
// line take precedence.
func ParseURL(url string) error {
    var resp, err = http.Get(url)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("%s: %s", url, resp.Status)
    }
    var data []byte
    data, err = io.ReadAll(resp.Body)
    if err != nil { return err }
    return parseConfig(url, string(data))
}
