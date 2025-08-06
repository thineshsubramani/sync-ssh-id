package ops

import (
	"strings"

	"github.com/thineshsubramani/sync-ssh-id/internal/config"
)

// escapeForSingleQuotes safe-escapes string for embedding as a single-quoted shell arg
func escapeForSingleQuotes(s string) string {
	if s == "" {
		return "''"
	}
	parts := strings.Split(s, "'")
	if len(parts) == 1 {
		return "'" + s + "'"
	}
	var b strings.Builder
	for i, p := range parts {
		if p != "" {
			b.WriteString("'" + p + "'")
		}
		if i != len(parts)-1 {
			b.WriteString(`'"'"'`)
		}
	}
	return b.String()
}

func addrAndPort(s config.Server) (host string, port string) {
	if s.IP != "" {
		host = s.IP
	} else {
		host = s.Host
	}
	port = strings.TrimSpace(s.Port)
	return
}
