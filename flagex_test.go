package flagex

import "flag"
import "os"
import "path/filepath"
import "testing"

var user = flag.String("user", "", "user name")
var url = flag.String("url", "http://localhost", "URL")
var listen = flag.String("listen", ":8080", "listen address")

func TestConfig(t *testing.T) {
    var path = filepath.Join(t.TempDir(), "test.conf")
    var err = os.WriteFile(path, []byte("# comment\n\nuser=configuser\nurl = http://localhost:8332\nlisten=:9999\n"), 0600)
    if err != nil {
        t.Fatal(err)
    }
    var oldUser, oldUrl, oldListen = *user, *url, *listen
    defer func() { *user = oldUser; *url = oldUrl; *listen = oldListen }()
    flag.Set("listen", ":7777")
    if err := ParseFile(path); err != nil {
        t.Fatalf("applyConfig: %v", err)
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
