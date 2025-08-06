package ops

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/thineshsubramani/sync-ssh-id/internal/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// privateKeyAuthMethods returns available local private key auth methods (publickeys)
func privateKeyAuthMethods() []ssh.AuthMethod {
	var out []ssh.AuthMethod
	usr, err := user.Current()
	if err != nil {
		return nil
	}
	cands := []string{
		filepath.Join(usr.HomeDir, ".ssh", "id_rsa"),
		filepath.Join(usr.HomeDir, ".ssh", "id_ed25519"),
		filepath.Join(usr.HomeDir, ".ssh", "id_ecdsa"),
	}
	for _, p := range cands {
		if _, err := os.Stat(p); err == nil {
			if signer, err := readPrivateKeySigner(p); err == nil {
				out = append(out, ssh.PublicKeys(signer))
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func readPrivateKeySigner(path string) (ssh.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if s, err := ssh.ParsePrivateKey(b); err == nil {
		return s, nil
	}
	return nil, fmt.Errorf("private key parse failed (maybe passphrase protected): %s", path)
}

// buildAuthMethods constructs available ssh.AuthMethod list and a short list of auth names for logging.
// Matches the logic used in the original dialer: password (per-host or env), ssh-agent, and local private keys.
func buildAuthMethods(s config.Server) ([]ssh.AuthMethod, []string) {
	var authMethods []ssh.AuthMethod
	var authNames []string

	// password from per-host config or global env
	pass := strings.TrimSpace(s.Pass)
	if pass == "" {
		pass = strings.TrimSpace(os.Getenv("SSH_PASS"))
	}
	if pass != "" {
		authMethods = append(authMethods, ssh.Password(pass))
		authMethods = append(authMethods, ssh.KeyboardInteractive(
			func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				for range questions {
					answers = append(answers, pass)
				}
				return answers, nil
			},
		))
		authNames = append(authNames, "password/keyboard-interactive")
	}

	// ssh-agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		if agConn, err := net.Dial("unix", sock); err == nil {
			agentClient := agent.NewClient(agConn)
			authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
			authNames = append(authNames, "agent")
			_ = agConn.Close()
		}
	}

	// local private keys (~/.ssh/*)
	if keyAuth := privateKeyAuthMethods(); keyAuth != nil {
		authMethods = append(authMethods, keyAuth...)
		authNames = append(authNames, "privatekey")
	}

	return authMethods, authNames
}
