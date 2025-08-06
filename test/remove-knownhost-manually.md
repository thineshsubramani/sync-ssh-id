## Reset SSH Host Key Fingerprint (for testing)

If you want to simulate **first-time SSH connection** behavior (no cached fingerprint), remove the host’s entry from `~/.ssh/known_hosts` before running the tool.

### Remove a specific host/IP

```bash
ssh-keygen -R <host-or-ip>
# Example:
ssh-keygen -R 192.168.100.99
```

### Remove using `sed` (alternative)

```bash
sed -i '/192\.168\.100\.99/d' ~/.ssh/known_hosts
```

### Wipe all known hosts (⚠ resets trust for ALL SSH connections)

```bash
> ~/.ssh/known_hosts
```

After resetting, your next SSH connection will prompt to trust the host key — useful for testing **first connection scenarios** in automation.
