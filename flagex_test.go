package flagex

import "flag"
import "net/http"
import "net/http/httptest"
import "os"
import "path/filepath"
import "strings"
import "testing"

var user = flag.String("user", "", "user name")
var url = flag.String("url", "http://localhost", "URL")
var listen = flag.String("listen", ":8080", "listen address")

var resetFlags = func() {
    flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
    user = flag.String("user", "", "user name")
    url = flag.String("url", "http://localhost", "URL")
    listen = flag.String("listen", ":8080", "listen address")
}

func TestParseFile(t *testing.T) {
    resetFlags()
    var path = filepath.Join(t.TempDir(), "test.conf")
    var err = os.WriteFile(path, []byte("# comment\n\nuser=configuser\nurl = http://localhost:8332\nlisten=:9999\n"), 0600)
    if err != nil {
        t.Fatal(err)
    }
    var oldUser, oldUrl, oldListen = *user, *url, *listen
    defer func() { *user = oldUser; *url = oldUrl; *listen = oldListen }()
    flag.Set("listen", ":7777")
    if err := ParseFile(path); err != nil {
        t.Fatalf("ParseFile: %v", err)
    }
    if *user != "configuser" {
        t.Fatalf("expected core-user from config, got %q", *user)
    }
    if *url != "http://localhost:8332" {
        t.Fatalf("expected core-url from config with spaces trimmed, got %q", *url)
    }
    if *listen != ":7777" {
        t.Fatalf("expected command-line listen to win over config, got %q", *listen)
    }
    if err := ParseFile(filepath.Join(t.TempDir(), "missing.conf")); err == nil {
        t.Fatalf("expected error for missing file")
    }
    os.WriteFile(path, []byte("unknown-option=1\n"), 0600)
    if err := ParseFile(path); err == nil {
        t.Fatalf("expected error for unknown option")
    }
    os.WriteFile(path, []byte("no equals sign\n"), 0600)
    if err := ParseFile(path); err == nil {
        t.Fatalf("expected error for malformed line")
    }
}

func TestParseStream(t *testing.T) {
    resetFlags()
    var oldUser, oldUrl, oldListen = *user, *url, *listen
    defer func() { *user = oldUser; *url = oldUrl; *listen = oldListen }()
    flag.Set("listen", ":7777")
    var r = strings.NewReader("user=streamuser\nurl = http://stream:9000\nlisten=:5555\n")
    if err := ParseStream(r); err != nil {
        t.Fatalf("ParseStream: %v", err)
    }
    if *user != "streamuser" {
        t.Fatalf("expected user from stream, got %q", *user)
    }
    if *url != "http://stream:9000" {
        t.Fatalf("expected url from stream, got %q", *url)
    }
    if *listen != ":7777" {
        t.Fatalf("expected command-line listen to win over stream, got %q", *listen)
    }
}

func TestParseStreamErrors(t *testing.T) {
    resetFlags()
    var oldUser = *user
    defer func() { *user = oldUser }()
    if err := ParseStream(strings.NewReader("unknown-option=1\n")); err == nil {
        t.Fatalf("expected error for unknown option")
    }
    if err := ParseStream(strings.NewReader("no equals sign\n")); err == nil {
        t.Fatalf("expected error for malformed line")
    }
    var errReader = &errorReader{msg: "test read error"}
    if err := ParseStream(errReader); err == nil {
        t.Fatalf("expected error for reader failure")
    }
}

func TestParseURL(t *testing.T) {
    resetFlags()
    var oldUser, oldUrl, oldListen = *user, *url, *listen
    defer func() { *user = oldUser; *url = oldUrl; *listen = oldListen }()
    flag.Set("listen", ":7777")
    var srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("user=urluser\nurl = http://remote:7000\nlisten=:8888\n"))
    }))
    defer srv.Close()
    if err := ParseURL(srv.URL); err != nil {
        t.Fatalf("ParseURL: %v", err)
    }
    if *user != "urluser" {
        t.Fatalf("expected user from URL, got %q", *user)
    }
    if *url != "http://remote:7000" {
        t.Fatalf("expected url from URL, got %q", *url)
    }
    if *listen != ":7777" {
        t.Fatalf("expected command-line listen to win over URL, got %q", *listen)
    }
}

func TestParseURLErrors(t *testing.T) {
    resetFlags()
    var oldUser = *user
    defer func() { *user = oldUser }()
    if err := ParseURL("http://127.0.0.1:1/nope"); err == nil {
        t.Fatalf("expected error for unreachable URL")
    }
    var srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()
    if err := ParseURL(srv.URL); err == nil {
        t.Fatalf("expected error for non-200 status")
    }
}

type errorReader struct{ msg string }

func (e *errorReader) Read(p []byte) (int, error) {
    return 0, &testError{e.msg}
}

type testError struct{ s string }

func (e *testError) Error() string { return e.s }
