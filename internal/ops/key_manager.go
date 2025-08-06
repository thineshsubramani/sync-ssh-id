package ops

import "time"

// KeyManager manages SSH key operations over SSH (native Go)
type KeyManager struct {
	DialTimeout time.Duration
}

func NewKeyManager() *KeyManager {
	return &KeyManager{DialTimeout: 15 * time.Second}
}
