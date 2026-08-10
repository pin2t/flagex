// Package flagex extends the standard flag package with functions to parse
// command-line flags from configuration files, streams, and URLs.
//
// Each line in the configuration should be in name=value format. Blank lines
// and lines starting with # are ignored. Flags already set via the command
// line take precedence and are not overwritten.
//
// The FlagSet type wraps a flag.FlagSet and provides the same configuration
// parsing for non-default FlagSets (e.g., those created with flag.NewFlagSet).
package flagex

import "flag"
import "fmt"
import "io"
import "net/http"
import "os"
import "strings"

// flagSet wraps a flag.FlagSet to provide configuration parsing from files,
// streams, URLs, and environment variables.
type flagSet struct {
	fs *flag.FlagSet
}

// FlagSet creates a new flagSet wrapper around fs. Use it to parse
// configuration for a specific flag.FlagSet instead of the default
// command-line FlagSet.
//
//	fs := flag.NewFlagSet("myapp", flag.ExitOnError)
//	flagex.FlagSet(fs).ParseFile("app.conf")
func FlagSet(fs *flag.FlagSet) *flagSet {
	return &flagSet{fs: fs}
}

// commandLine returns a flagSet wrapping flag.CommandLine.
func commandLine() *flagSet {
	return &flagSet{fs: flag.CommandLine}
}

// parseConfig is the internal parsing engine shared by all Parse* methods.
func (fs *flagSet) parseConfig(source string, data string) error {
	var set = make(map[string]bool)
	fs.fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	for i, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") { continue }
		var name, value, found = strings.Cut(line, "=")
		if !found {
			return fmt.Errorf("%s:%d: expected name=value, got %q", source, i+1, line)
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if fs.fs.Lookup(name) == nil {
			return fmt.Errorf("%s:%d: unknown option %q", source, i+1, name)
		}
		if set[name] { continue }
		if err := fs.fs.Set(name, value); err != nil {
			return fmt.Errorf("%s:%d: %v", source, i+1, err)
		}
	}
	return nil
}

// ParseFile reads a configuration file at path and sets any flags in the
// FlagSet that have not already been set.
func (fs *flagSet) ParseFile(path string) error {
	var data, err = os.ReadFile(path)
	if err != nil { return err }
	return fs.parseConfig(path, string(data))
}

// ParseStream reads configuration from r and sets flags in the FlagSet.
// Flags already set take precedence and are not overwritten.
func (fs *flagSet) ParseStream(r io.Reader) error {
	var data, err = io.ReadAll(r)
	if err != nil { return err }
	return fs.parseConfig("stream", string(data))
}

// ParseURL fetches configuration from url via HTTP GET and sets flags in the
// FlagSet. A non-200 status code returns an error. Flags already set take
// precedence.
func (fs *flagSet) ParseURL(url string) error {
	var resp, err = http.Get(url)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	var data []byte
	data, err = io.ReadAll(resp.Body)
	if err != nil { return err }
	return fs.parseConfig(url, string(data))
}

// ParseEnv reads environment variables and sets matching flags in the
// FlagSet. Environment variable names are mapped to flag names by stripping
// prefix and converting to lowercase. For example, with prefix "APP_",
// APP_URL sets the "url" flag.
//
// When prefix is non-empty, only variables starting with the prefix are
// processed and any that do not match a defined flag cause an error.
// When prefix is empty, all variables are considered and those that do
// not match a defined flag are silently skipped.
// Flags already set take precedence.
func (fs *flagSet) ParseEnv(prefix string) error {
	var set = make(map[string]bool)
	fs.fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	for _, kv := range os.Environ() {
		var name, value, found = strings.Cut(kv, "=")
		if !found { continue }
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		var flagName string
		if prefix == "" {
			flagName = strings.ToLower(name)
		} else {
			if !strings.HasPrefix(name, prefix) { continue }
			flagName = strings.ToLower(strings.TrimPrefix(name, prefix))
			if flagName == "" { continue }
		}
		if fs.fs.Lookup(flagName) == nil {
			if prefix != "" {
				return fmt.Errorf("env %s: unknown option %q", name, flagName)
			}
			continue
		}
		if set[flagName] { continue }
		if err := fs.fs.Set(flagName, value); err != nil {
			return fmt.Errorf("env %s: %v", name, err)
		}
	}
	return nil
}

// --- Package-level convenience functions ---
// These delegate to defaultFlagSet wrapping flag.CommandLine.

// ParseFile reads a configuration file at path and sets any flags defined in the
// standard flag package that have not already been set on the command line.
func ParseFile(path string) error {
	return commandLine().ParseFile(path)
}

// ParseStream reads configuration from r and sets flags in the standard flag
// package. Flags already set via the command line are not overwritten.
func ParseStream(r io.Reader) error {
	return commandLine().ParseStream(r)
}

// ParseURL fetches configuration from url via HTTP GET and sets flags.
// A non-200 status code returns an error. Flags already set on the command
// line take precedence.
func ParseURL(url string) error {
	return commandLine().ParseURL(url)
}

// ParseEnv reads environment variables and sets matching flags in the
// standard flag package. Environment variable names are mapped to flag
// names by stripping prefix and converting to lowercase. For example,
// with prefix "APP_", APP_URL sets the "url" flag.
//
// When prefix is non-empty, only variables starting with the prefix are
// processed and any that do not match a defined flag cause an error.
// When prefix is empty, all variables are considered and those that do
// not match a defined flag are silently skipped.
// Flags already set via the command line take precedence.
func ParseEnv(prefix string) error {
	return commandLine().ParseEnv(prefix)
}
