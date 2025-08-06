package util

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandPath expands only "~/" and env vars, leaves rest untouched
func ExpandPath(p string) string {
	if p == "" {
		return ""
	}

	// Expand environment variables
	p = os.ExpandEnv(p)

	// Expand ~/ only
	if p == "~" || strings.HasPrefix(p, "~/") {
		if usr, err := user.Current(); err == nil {
			p = filepath.Join(usr.HomeDir, p[1:])
		}
	}

	return p
}

// CommandExists checks if command exists in PATH
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// KnownHostsPath returns default known_hosts path for current user
func KnownHostsPath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".ssh", "known_hosts")
}
