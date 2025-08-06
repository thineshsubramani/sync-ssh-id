package env

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

// SmartEnvMap searches for best matching env file and returns the map and path
func SmartEnvMap(envDir, host, ip string) (map[string]string, string, error) {
	candidates := []string{
		fmt.Sprintf(".env.%s", host),
		fmt.Sprintf(".env.%s", ip),
		fmt.Sprintf(".env.%s", sanitizeHost(host)),
		".env.default",
	}
	for _, c := range candidates {
		p := filepath.Join(envDir, c)
		if _, err := os.Stat(p); err == nil {
			m, err := godotenv.Read(p)
			return m, p, err
		}
	}
	return map[string]string{}, "", nil
}

func sanitizeHost(h string) string {
	if h == "" {
		return ""
	}
	// take first label and strip non-alnum/- characters
	parts := strings.Split(h, ".")
	label := parts[0]
	re := regexp.MustCompile(`[^a-zA-Z0-9\-]`)
	return re.ReplaceAllString(label, "")
}
