package cli

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thineshsubramani/sync-ssh-id/internal/config"
	"github.com/thineshsubramani/sync-ssh-id/internal/env"
	"github.com/thineshsubramani/sync-ssh-id/internal/ops"
	"github.com/thineshsubramani/sync-ssh-id/internal/output"
)

// defaultRemotePath returns the default remote authorized_keys path
func defaultRemotePath() string {
	return "~/.ssh/authorized_keys"
}

// RunInventory processes inventory YAML, rendering each server separately with its env map.
// It prints only one final status line per server (OK or ERROR).
func RunInventory(opts *Options) error {
	raw, err := config.LoadRaw(opts.ConfigPath)
	if err != nil {
		return err
	}

	inv, err := config.ParseInventory(raw)
	if err != nil {
		return err
	}

	for _, srv := range inv.Servers {
		// get per-host env map (does NOT mutate process env)
		envMap, _, _ := env.SmartEnvMap(opts.EnvDir, srv.Host, srv.IP)

		// marshal only this server so templating applies to just this block
		smallInv := config.Inventory{Servers: []config.Server{srv}}
		smallRaw, err := yaml.Marshal(&smallInv)
		if err != nil {
			return err
		}

		// render only the small YAML with envMap (no cross-talk)
		rendered, err := config.RenderForHost(smallRaw, envMap)
		if err != nil {
			return err
		}

		hostInv, err := config.ParseInventory(rendered)
		if err != nil {
			return err
		}

		// should be exactly one server here
		for _, h := range hostInv.Servers {
			// PASSWORD PRECEDENCE:
			// 1) explicit short form: h.Pass (yaml: pass)
			// 2) explicit full form: h.Password (yaml: password)
			// 3) per-host env: envMap["SSH_PASS"]
			// 4) global env: os.Getenv("SSH_PASS")
			if strings.TrimSpace(h.Pass) == "" {
				if strings.TrimSpace(h.Password) != "" {
					h.Pass = strings.TrimSpace(h.Password)
				} else if v, ok := envMap["SSH_PASS"]; ok && strings.TrimSpace(v) != "" {
					h.Pass = strings.TrimSpace(v)
				} else if global := os.Getenv("SSH_PASS"); strings.TrimSpace(global) != "" {
					h.Pass = strings.TrimSpace(global)
				}
			}

			// PUBLIC KEY PRECEDENCE:
			// 1) explicit in YAML h.PublicKey
			// 2) envMap["PUB_KEY_PATH"]
			// 3) global PUB_KEY_PATH env
			if strings.TrimSpace(h.PublicKey) == "" {
				if v, ok := envMap["PUB_KEY_PATH"]; ok && strings.TrimSpace(v) != "" {
					h.PublicKey = strings.TrimSpace(v)
				} else if global := os.Getenv("PUB_KEY_PATH"); strings.TrimSpace(global) != "" {
					h.PublicKey = strings.TrimSpace(global)
				}
			}

			remotePath := defaultRemotePath()

			// If dry-run: don't perform actions, just print OK once
			if opts.DryRun {
				output.OK(h, remotePath)
				continue
			}

			mgr := ops.NewKeyManager()
			action := strings.ToLower(h.Action)
			switch action {
			case "inject", "add":
				if err := mgr.Inject(h); err != nil {
					output.Error(h, remotePath, err)
				} else {
					output.OK(h, remotePath)
				}
			case "delete", "remove":
				if err := mgr.Delete(h); err != nil {
					output.Error(h, remotePath, err)
				} else {
					output.OK(h, remotePath)
				}
			case "update":
				if err := mgr.Update(h); err != nil {
					output.Error(h, remotePath, err)
				} else {
					output.OK(h, remotePath)
				}
			default:
				output.Error(h, remotePath, nil)
			}
		}
	}
	return nil
}
