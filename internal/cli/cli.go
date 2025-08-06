package cli

import (
	"flag"
	"os"
)

type Options struct {
	Interactive bool
	EnvDir      string
	DryRun      bool
	ConfigPath  string
	Args        []string // leftover args

	Host       string // in interactive mode can be user@host
	Pass       string
	PubKey     string
	RemotePath string
	Port       string
}

func ParseFlags() *Options {
	opts := &Options{}

	flag.BoolVar(&opts.Interactive, "i", false, "interactive single-host mode")
	flag.StringVar(&opts.EnvDir, "env-dir", "configs", "directory to search env files")
	flag.BoolVar(&opts.DryRun, "dry-run", false, "dry run - no changes")

	flag.StringVar(&opts.Host, "host", "", "target hostname or IP (can be user@host)")
	flag.StringVar(&opts.Pass, "pass", "", "remote password")
	// keep flag names compatible with earlier examples (local_path / remote_path)
	flag.StringVar(&opts.PubKey, "local_path", "", "local public key path")
	flag.StringVar(&opts.RemotePath, "remote_path", "", "remote authorized_keys path")
	flag.StringVar(&opts.Port, "port", "", "ssh port for interactive mode (optional)")

	flag.Parse()
	opts.Args = flag.Args()

	if !opts.Interactive && len(opts.Args) > 0 {
		opts.ConfigPath = opts.Args[0]
	} else if !opts.Interactive && len(opts.Args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	return opts
}
