package config

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Name      string `yaml:"name"`
	Host      string `yaml:"host"`
	IP        string `yaml:"ip"`
	User      string `yaml:"user"`
	Port      string `yaml:"port"`
	PublicKey string `yaml:"public_key"`
	Action    string `yaml:"action"`   // inject|delete|update
	Pass      string `yaml:"pass"`     // legacy / short form
	Password  string `yaml:"password"` // support full 'password:' key in YAML
}

type Options struct {
	ResetKnownHost       bool `yaml:"reset_knownhost"`
	BackupAuthorizedKeys bool `yaml:"backup_authorized_keys"`
}

type Inventory struct {
	Servers []Server `yaml:"servers"`
	Options Options  `yaml:"options"`
}

// LoadRaw returns raw bytes of config file
func LoadRaw(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// It exposes each key in vars as a template function so templates like
// `{{SSH_USER}}` will resolve correctly. missingkey=error ensures templates
// fail loudly if a required var is missing.
func RenderForHost(raw []byte, vars map[string]string) ([]byte, error) {
	// Build FuncMap where each env key is a zero-arg function returning its value.
	funcs := template.FuncMap{}
	for k, v := range vars {
		val := v // capture loop var
		funcs[k] = func() string { return val }
	}

	tmpl := template.New("cfg").Funcs(funcs).Option("missingkey=error")

	t, err := tmpl.Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	// Execute with nil data because vars are available as functions.
	if err := t.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}

	return buf.Bytes(), nil
}

// ParseInventory parses YAML bytes into Inventory struct
func ParseInventory(raw []byte) (*Inventory, error) {
	var inv Inventory
	if err := yaml.Unmarshal(raw, &inv); err != nil {
		return nil, err
	}
	return &inv, nil
}
