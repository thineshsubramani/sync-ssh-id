package ops

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/thineshsubramani/sync-ssh-id/internal/config"
	"github.com/thineshsubramani/sync-ssh-id/internal/util"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// dialForServer returns ssh.Client for given server, using s.User, s.Port, s.IP/Host, and password if present.
// It will auto-add the host key to known_hosts on *first-time connection* (only when host key is missing),
// but will NOT auto-accept mismatched host keys.
func (k *KeyManager) dialForServer(s config.Server) (*ssh.Client, error) {
	host, port := addrAndPort(s)
	if port == "" {
		port = "22"
	}

	// build auth methods
	authMethods, _ := buildAuthMethods(s)

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no auth methods available for host=%s user=%s", host, s.User)
	}

	// Build known_hosts callback if possible
	khPath := util.KnownHostsPath()
	var knownCb ssh.HostKeyCallback
	if cb, err := knownhosts.New(khPath); err == nil {
		knownCb = cb
	} else {
		// no known_hosts file or unreadable — fall back to insecure for now (we'll capture later)
		knownCb = ssh.InsecureIgnoreHostKey()
	}

	// create initial config that uses known_hosts verification
	cfg := &ssh.ClientConfig{
		User:            s.User,
		Auth:            authMethods,
		HostKeyCallback: knownCb,
		Timeout:         k.DialTimeout,
	}

	addr := net.JoinHostPort(host, port)
	client, err := ssh.Dial("tcp", addr, cfg)
	if err == nil {
		// log.Printf("[%s] auth methods: %s", addr, strings.Join(authNames, ","))   #
		return client, nil
	}

	// If error is HostKeyError from knownhosts, we may be missing an entry (first-time).
	var hkErr *knownhosts.KeyError
	if errors.As(err, &hkErr) {
		// If there are no "Want" entries then host not present in known_hosts (first-time).
		if len(hkErr.Want) == 0 {
			log.Printf("[%s] host key not found in known_hosts, auto-fetching and adding entry", host)

			capturedKey, tmpErr := fetchAndCaptureHostKey(addr, s.User, authMethods, k.DialTimeout)
			if tmpErr != nil {
				return nil, fmt.Errorf("temp fetch hostkey failed: %w", tmpErr)
			}
			if capturedKey == nil {
				return nil, fmt.Errorf("failed to capture remote host key for %s", addr)
			}

			// build known_hosts line and append to file
			line := knownhosts.Line([]string{host}, capturedKey)
			f, ferr := os.OpenFile(khPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
			if ferr != nil {
				return nil, fmt.Errorf("cannot open known_hosts (%s): %w", khPath, ferr)
			}
			if _, ferr = f.WriteString(line + "\n"); ferr != nil {
				_ = f.Close()
				return nil, fmt.Errorf("failed writing known_hosts: %w", ferr)
			}
			_ = f.Close()

			// Retry original dial with verified callback
			if cb, err := knownhosts.New(khPath); err == nil {
				cfg.HostKeyCallback = cb
			} else {
				cfg.HostKeyCallback = ssh.InsecureIgnoreHostKey()
			}

			client2, err2 := ssh.Dial("tcp", addr, cfg)
			if err2 != nil {
				return nil, fmt.Errorf("ssh dial after adding known_hosts failed: %w", err2)
			}
			log.Printf("[%s] added host key to known_hosts and connected", addr)
			return client2, nil
		}

		// if Want is non-empty that's a mismatch (host key changed) — do NOT auto-accept
		return nil, fmt.Errorf("knownhosts: key mismatch for %s: %v", host, hkErr)
	}

	// not a knownhosts error — return original error
	return nil, fmt.Errorf("ssh dial to %s failed: %w", addr, err)
}

// fetchAndCaptureHostKey makes a temporary SSH connection that accepts the remote host key and returns that key for persistence.
func fetchAndCaptureHostKey(addr, user string, authMethods []ssh.AuthMethod, timeout time.Duration) (ssh.PublicKey, error) {
	var capturedKey ssh.PublicKey
	tempCfg := &ssh.ClientConfig{
		User: user,
		Auth: authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			capturedKey = key
			return nil // accept for this temp connection
		},
		Timeout: timeout,
	}

	tmpClient, tmpErr := ssh.Dial("tcp", addr, tempCfg)
	if tmpErr != nil {
		return nil, tmpErr
	}
	_ = tmpClient.Close()
	return capturedKey, nil
}
