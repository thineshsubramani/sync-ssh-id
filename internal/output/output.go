package output

import (
	"log"
	"time"

	"github.com/thineshsubramani/sync-ssh-id/internal/config"
	"github.com/thineshsubramani/sync-ssh-id/internal/util"
)

type Status string

const (
	StatusStart Status = "START"
	StatusOK    Status = "OK"
	StatusError Status = "ERROR"
)

// ANSI colors
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

type LogEntry struct {
	Host       string
	Action     string
	User       string
	Pass       string
	PubKey     string
	RemotePath string
	Status     Status
	Err        error
}

func OK(h config.Server, remotePath string) {
	Print(LogEntry{
		Host:       h.Host,
		Action:     h.Action,
		User:       h.User,
		Pass:       h.Pass,
		PubKey:     h.PublicKey,
		RemotePath: remotePath,
		Status:     StatusOK,
	})
}

func Error(h config.Server, remotePath string, err error) {
	Print(LogEntry{
		Host:       h.Host,
		Action:     h.Action,
		User:       h.User,
		Pass:       h.Pass,
		PubKey:     h.PublicKey,
		RemotePath: remotePath,
		Status:     StatusError,
		Err:        err,
	})
}

func Print(entry LogEntry) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	passMasked := util.Mask(entry.Pass)

	// Colorize status
	var statusColored string
	switch entry.Status {
	case StatusOK:
		statusColored = colorGreen + "Success" + colorReset
	case StatusError:
		statusColored = colorRed + "ERROR" + colorReset
	case StatusStart:
		statusColored = colorYellow + "START" + colorReset
	}

	// Build log line
	if entry.Status == StatusError && entry.Err != nil {
		log.Printf("%s  %-15s  %-8s  %-10s  %-10s  %-30s  %-30s  %s (%v)",
			timestamp,
			entry.Host,
			entry.Action,
			entry.User,
			passMasked,
			entry.PubKey,
			entry.RemotePath,
			statusColored,
			entry.Err,
		)
	} else {
		log.Printf("%s  %-15s  %-8s  %-10s  %-10s  %-30s  %-30s  %s",
			timestamp,
			entry.Host,
			entry.Action,
			entry.User,
			passMasked,
			entry.PubKey,
			entry.RemotePath,
			statusColored,
		)
	}
}
