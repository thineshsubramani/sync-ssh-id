package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/thineshsubramani/sync-ssh-id/internal/config"
	"github.com/thineshsubramani/sync-ssh-id/internal/ops"
	"github.com/thineshsubramani/sync-ssh-id/internal/output"
	"github.com/thineshsubramani/sync-ssh-id/internal/util"
)

// RunInteractive runs single-host interactive flow like ssh-copy-id
func RunInteractive(opts *Options) error {
	target := strings.TrimSpace(opts.Host)
	if target == "" && len(opts.Args) > 0 {
		target = opts.Args[0] // allow "user@host" as positional arg
	}

	if target == "" || !strings.Contains(target, "@") {
		return fmt.Errorf("must specify target as user@host (e.g., -i user@1.2.3.4)")
	}

	parts := strings.SplitN(target, "@", 2)
	user := parts[0]
	host := parts[1]

	// auto-detect key paths
	if strings.TrimSpace(opts.PubKey) == "" {
		opts.PubKey = util.DetectDefaultPubKey()
	}
	if strings.TrimSpace(opts.RemotePath) == "" {
		opts.RemotePath = "~/.ssh/authorized_keys"
	}

	// prompt for password if not given
	if strings.TrimSpace(opts.Pass) == "" {
		fmt.Printf("[%s@%s] Password: ", user, host)
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		opts.Pass = strings.TrimSpace(string(pw))
	}

	// expand and clean local pubkey path
	pubkeyPath := filepath.Clean(os.ExpandEnv(opts.PubKey))
	if _, err := os.Stat(pubkeyPath); err != nil {
		return fmt.Errorf("public key not found at %s", pubkeyPath)
	}

	// load public key file
	pubData, err := os.ReadFile(pubkeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	h := config.Server{
		Name:      host,
		Host:      host,
		User:      user,
		PublicKey: pubkeyPath,
		Action:    "inject",
		Pass:      opts.Pass,
		Port:      opts.Port,
	}

	remotePath := strings.TrimSpace(opts.RemotePath)

	if opts.DryRun {
		output.OK(h, remotePath)
		return nil
	}

	mgr := ops.NewKeyManager()
	if err := mgr.InjectWithCustomPath(h, string(pubData), remotePath); err != nil {
		output.Error(h, remotePath, err)
		return err
	}

	output.OK(h, remotePath)
	return nil
}
