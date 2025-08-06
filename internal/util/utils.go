package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func DetectDefaultPubKey() string {
	if v := os.Getenv("PUB_KEY_PATH"); v != "" {
		return v
	}
	if runtime.GOOS == "windows" {
		if up := os.Getenv("USERPROFILE"); up != "" {
			return filepath.Join(up, ".ssh", "id_rsa.pub")
		}
		return filepath.Join(os.Getenv("HOMEDRIVE")+os.Getenv("HOMEPATH"), ".ssh", "id_rsa.pub")
	}
	if u, err := os.UserHomeDir(); err == nil {
		return filepath.Join(u, ".ssh", "id_rsa.pub")
	}
	return "./.ssh/id_rsa.pub"
}

func Mask(s string) string {
	if s == "" {
		return "<empty>"
	}
	if len(s) <= 2 {
		return "**"
	}
	return s[:1] + strings.Repeat("*", len(s)-2) + s[len(s)-1:]
}
