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
var count = flag.Int("count", 0, "count value")
var verbose = flag.Bool("verbose", false, "verbose mode")
var ratio = flag.Float64("ratio", 1.0, "ratio value")

var resetFlags = func() {
    flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
    user = flag.String("user", "", "user name")
    url = flag.String("url", "http://localhost", "URL")
    listen = flag.String("listen", ":8080", "listen address")
    count = flag.Int("count", 0, "count value")
    verbose = flag.Bool("verbose", false, "verbose mode")
    ratio = flag.Float64("ratio", 1.0, "ratio value")
}

func TestParseFile(t *testing.T) {
    resetFlags()
    var path = filepath.Join(t.TempDir(), "test.conf")
    var err = os.WriteFile(path, []byte("# comment\n\nuser=configuser\nurl = http://localhost:8332\nlisten=:9999\ncount=42\nverbose=true\nratio=0.75\n"), 0600)
    if err != nil {
        t.Fatal(err)
    }
    var oldUser, oldUrl, oldListen, oldCount, oldVerbose, oldRatio = *user, *url, *listen, *count, *verbose, *ratio
    defer func() { *user = oldUser; *url = oldUrl; *listen = oldListen; *count = oldCount; *verbose = oldVerbose; *ratio = oldRatio }()
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
    if *count != 42 {
        t.Fatalf("expected count=42 from config, got %d", *count)
    }
    if *verbose != true {
        t.Fatalf("expected verbose=true from config, got %v", *verbose)
    }
    if *ratio != 0.75 {
        t.Fatalf("expected ratio=0.75 from config, got %v", *ratio)
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
    var oldUser, oldUrl, oldListen, oldCount, oldVerbose, oldRatio = *user, *url, *listen, *count, *verbose, *ratio
    defer func() { *user = oldUser; *url = oldUrl; *listen = oldListen; *count = oldCount; *verbose = oldVerbose; *ratio = oldRatio }()
    flag.Set("listen", ":7777")
    var r = strings.NewReader("user=streamuser\nurl = http://stream:9000\nlisten=:5555\ncount=10\nverbose=false\nratio=2.5\n")
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
    if *count != 10 {
        t.Fatalf("expected count=10 from stream, got %d", *count)
    }
    if *verbose != false {
        t.Fatalf("expected verbose=false from stream, got %v", *verbose)
    }
    if *ratio != 2.5 {
        t.Fatalf("expected ratio=2.5 from stream, got %v", *ratio)
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
    var oldUser, oldUrl, oldListen, oldCount, oldVerbose, oldRatio = *user, *url, *listen, *count, *verbose, *ratio
    defer func() { *user = oldUser; *url = oldUrl; *listen = oldListen; *count = oldCount; *verbose = oldVerbose; *ratio = oldRatio }()
    flag.Set("listen", ":7777")
    var srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("user=urluser\nurl = http://remote:7000\nlisten=:8888\ncount=7\nverbose=true\nratio=3.14\n"))
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
    if *count != 7 {
        t.Fatalf("expected count=7 from URL, got %d", *count)
    }
    if *verbose != true {
        t.Fatalf("expected verbose=true from URL, got %v", *verbose)
    }
    if *ratio != 3.14 {
        t.Fatalf("expected ratio=3.14 from URL, got %v", *ratio)
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

func TestParseEnvWithPrefix(t *testing.T) {
    resetFlags()
    var oldUser, oldUrl, oldListen, oldCount, oldVerbose, oldRatio = *user, *url, *listen, *count, *verbose, *ratio
    defer func() { *user = oldUser; *url = oldUrl; *listen = oldListen; *count = oldCount; *verbose = oldVerbose; *ratio = oldRatio }()
    t.Setenv("APP_USER", "envuser")
    t.Setenv("APP_URL", "http://env:9000")
    t.Setenv("APP_LISTEN", ":8888")
    t.Setenv("APP_COUNT", "99")
    t.Setenv("APP_VERBOSE", "true")
    t.Setenv("APP_RATIO", "1.5")
    if err := ParseEnv("APP_"); err != nil {
        t.Fatalf("ParseEnv: %v", err)
    }
    if *user != "envuser" {
        t.Fatalf("expected user from env, got %q", *user)
    }
    if *url != "http://env:9000" {
        t.Fatalf("expected url from env, got %q", *url)
    }
    if *listen != ":8888" {
        t.Fatalf("expected listen from env, got %q", *listen)
    }
    if *count != 99 {
        t.Fatalf("expected count=99 from env, got %d", *count)
    }
    if *verbose != true {
        t.Fatalf("expected verbose=true from env, got %v", *verbose)
    }
    if *ratio != 1.5 {
        t.Fatalf("expected ratio=1.5 from env, got %v", *ratio)
    }
}

func TestParseEnvPrecedence(t *testing.T) {
    resetFlags()
    var oldListen = *listen
    defer func() { *listen = oldListen }()
    flag.Set("listen", ":7777")
    t.Setenv("APP_LISTEN", ":8888")
    if err := ParseEnv("APP_"); err != nil {
        t.Fatalf("ParseEnv: %v", err)
    }
    if *listen != ":7777" {
        t.Fatalf("expected command-line listen to win over env, got %q", *listen)
    }
}

func TestParseEnvErrors(t *testing.T) {
    resetFlags()
    var oldUser = *user
    defer func() { *user = oldUser }()
    t.Setenv("APP_FOO", "bar")
    if err := ParseEnv("APP_"); err == nil {
        t.Fatalf("expected error for unknown option with prefix")
    }
}

func TestParseEnvInvalidValue(t *testing.T) {
    resetFlags()
    var oldCount = *count
    defer func() { *count = oldCount }()
    t.Setenv("APP_COUNT", "notanint")
    if err := ParseEnv("APP_"); err == nil {
        t.Fatalf("expected error for invalid value")
    }
}

func TestParseEnvNoPrefix(t *testing.T) {
    resetFlags()
    var oldUser, oldUrl = *user, *url
    defer func() { *user = oldUser; *url = oldUrl }()
    t.Setenv("USER", "envuser")
    t.Setenv("URL", "http://env:9000")
    t.Setenv("SOME_RANDOM_VAR", "should-be-ignored")
    if err := ParseEnv(""); err != nil {
        t.Fatalf("ParseEnv: %v", err)
    }
    if *user != "envuser" {
        t.Fatalf("expected user from env, got %q", *user)
    }
    if *url != "http://env:9000" {
        t.Fatalf("expected url from env, got %q", *url)
    }
    if *listen != ":8080" {
        t.Fatalf("expected default listen value, got %q", *listen)
    }
}

func TestParseEnvPrefixSkipsUnrelated(t *testing.T) {
    resetFlags()
    var oldUser = *user
    defer func() { *user = oldUser }()
    t.Setenv("OTHER_USER", "other")
    t.Setenv("APP_USER", "envuser")
    if err := ParseEnv("APP_"); err != nil {
        t.Fatalf("ParseEnv: %v", err)
    }
    if *user != "envuser" {
        t.Fatalf("expected user from APP_USER, got %q", *user)
    }
}
