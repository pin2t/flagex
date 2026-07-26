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

func ParseFile(path string) error {
    var data, err = os.ReadFile(path)
    if err != nil { return err }
    return parseConfig(path, string(data))
}

func ParseStream(r io.Reader) error {
    var data, err = io.ReadAll(r)
    if err != nil { return err }
    return parseConfig("stream", string(data))
}

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
