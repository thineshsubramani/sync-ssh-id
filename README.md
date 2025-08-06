# sync-ssh-id

Lightweight CLI to inject/update/delete SSH public keys across a fleet.
Supports per-host `.env` overrides, templated YAML inventory, and known_hosts fingerprint handling.

Quick start:
1. Edit `configs/example.yaml` and per-host `.env.<host>`.
2. Build: `go build ./...`
3. Run dry-run:
   `go run ./cmd/sync-ssh-id --dry-run configs/example.yaml`
