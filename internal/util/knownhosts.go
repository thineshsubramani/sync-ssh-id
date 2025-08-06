package util

import (
	"os"
	"strings"
)

// ResetKnownHost removes lines matching host from known_hosts
func ResetKnownHost(host string) error {
	path := KnownHostsPath()
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	var keep []string
	for _, l := range lines {
		if l == "" {
			continue
		}
		if strings.Contains(l, host) {
			continue
		}
		keep = append(keep, l)
	}
	return os.WriteFile(path, []byte(strings.Join(keep, "\n")), 0644)
}
