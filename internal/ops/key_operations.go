package ops

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/thineshsubramani/sync-ssh-id/internal/config"
	"github.com/thineshsubramani/sync-ssh-id/internal/util"
)

// Inject appends public key to remote authorized_keys idempotently
func (k *KeyManager) Inject(s config.Server) error {
	pubPath := util.ExpandPath(s.PublicKey)
	if pubPath == "" {
		usr, _ := user.Current()
		pubPath = filepath.Join(usr.HomeDir, ".ssh", "id_rsa.pub")
	}
	if _, err := os.Stat(pubPath); err != nil {
		return fmt.Errorf("public key not found: %s", pubPath)
	}

	keyData, err := os.ReadFile(pubPath)
	if err != nil {
		return fmt.Errorf("read pubkey: %w", err)
	}
	pubLine := strings.TrimSpace(string(keyData))

	client, err := k.dialForServer(s)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	if err := k.ensureSSHDir(client); err != nil {
		return fmt.Errorf("ensure ssh dir: %w", err)
	}

	if err := k.backupAuthorizedKeys(client); err != nil {
		return fmt.Errorf("backup authorized_keys: %w", err)
	}

	exists, err := k.remoteHasExactLine(client, pubLine)
	if err != nil {
		return fmt.Errorf("check existing key: %w", err)
	}
	if exists {
		return nil
	}

	if err := k.remoteAppendLine(client, pubLine); err != nil {
		return fmt.Errorf("append pubkey: %w", err)
	}

	if err := k.remoteChmod(client, "~/.ssh/authorized_keys", "600"); err != nil {
		return fmt.Errorf("set perms: %w", err)
	}

	return nil
}

// InjectWithCustomPath injects given pubKey into custom remotePath
func (k *KeyManager) InjectWithCustomPath(s config.Server, pubKey string, remotePath string) error {
	pub := strings.TrimSpace(pubKey)
	if pub == "" {
		return fmt.Errorf("public key content is empty")
	}
	if strings.TrimSpace(remotePath) == "" {
		remotePath = "~/.ssh/authorized_keys"
	}

	client, err := k.dialForServer(s)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	escaped := escapeForSingleQuotes(pub)

	cmd := fmt.Sprintf(
		"mkdir -p $(dirname %s) && touch %s && grep -F -x %s %s >/dev/null 2>&1 || printf '%%s\\n' %s >> %s && chmod 600 %s",
		remotePath, remotePath, escaped, remotePath, escaped, remotePath, remotePath,
	)

	out, err := runRemote(client, cmd)
	if err != nil {
		return fmt.Errorf("remote cmd failed: %v, output: %s", err, out)
	}

	return nil
}

// Delete removes the public key from remote authorized_keys
func (k *KeyManager) Delete(s config.Server) error {
	pubPath := util.ExpandPath(s.PublicKey)
	if pubPath == "" {
		usr, _ := user.Current()
		pubPath = filepath.Join(usr.HomeDir, ".ssh", "id_rsa.pub")
	}
	if _, err := os.Stat(pubPath); err != nil {
		return fmt.Errorf("public key not found: %s", pubPath)
	}
	keyData, err := os.ReadFile(pubPath)
	if err != nil {
		return fmt.Errorf("read pubkey: %w", err)
	}
	pubLine := strings.TrimSpace(string(keyData))

	client, err := k.dialForServer(s)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	if err := k.backupAuthorizedKeys(client); err != nil {
		return fmt.Errorf("backup authorized_keys: %w", err)
	}

	if err := k.remoteRemoveExactLine(client, pubLine); err != nil {
		return fmt.Errorf("remove pubkey: %w", err)
	}
	return nil
}

// Update removes then reinjects the public key
func (k *KeyManager) Update(s config.Server) error {
	_ = k.Delete(s)
	return k.Inject(s)
}
