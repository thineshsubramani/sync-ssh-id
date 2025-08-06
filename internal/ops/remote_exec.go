package ops

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// runRemote runs a single shell command on the remote client and returns combined output.
func runRemote(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &b
	if err := session.Run(cmd); err != nil {
		return b.String(), err
	}
	return b.String(), nil
}

func (k *KeyManager) ensureSSHDir(client *ssh.Client) error {
	_, err := runRemote(client, "mkdir -p ~/.ssh && chmod 700 ~/.ssh")
	return err
}

func (k *KeyManager) backupAuthorizedKeys(client *ssh.Client) error {
	ts := time.Now().UTC().Format("20060102T150405Z")
	_, err := runRemote(client, fmt.Sprintf("if [ -f ~/.ssh/authorized_keys ]; then cp ~/.ssh/authorized_keys ~/.ssh/authorized_keys.bak.%s; fi", ts))
	return err
}

func (k *KeyManager) remoteHasExactLine(client *ssh.Client, line string) (bool, error) {
	escaped := escapeForSingleQuotes(line)
	cmd := fmt.Sprintf("if [ -f ~/.ssh/authorized_keys ]; then grep -F -x %s ~/.ssh/authorized_keys >/dev/null 2>&1 && echo present || echo absent; else echo absent; fi", escaped)
	out, _ := runRemote(client, cmd)
	return strings.TrimSpace(out) == "present", nil
}

func (k *KeyManager) remoteAppendLine(client *ssh.Client, line string) error {
	escaped := escapeForSingleQuotes(line)
	_, err := runRemote(client, fmt.Sprintf("printf '%%s\\n' %s >> ~/.ssh/authorized_keys", escaped))
	return err
}

// remoteRemoveExactLine removes exact matching lines from authorized_keys
func (k *KeyManager) remoteRemoveExactLine(client *ssh.Client, line string) error {
	escaped := escapeForSingleQuotes(line)
	// Ensure grep failure (exit 1) doesn't stop the pipeline; redirect will still create tmp file.
	// Using '|| true' ensures the shell command exits 0, so SSH session.Run sees success.
	cmd := fmt.Sprintf("if [ -f ~/.ssh/authorized_keys ]; then grep -v -F -x %s ~/.ssh/authorized_keys > /tmp/authorized_keys.tmp || true; mv /tmp/authorized_keys.tmp ~/.ssh/authorized_keys; fi", escaped)
	_, err := runRemote(client, cmd)
	return err
}

func (k *KeyManager) remoteChmod(client *ssh.Client, path, mode string) error {
	_, err := runRemote(client, fmt.Sprintf("chmod %s %s", mode, path))
	return err
}
