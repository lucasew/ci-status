package config

import (
	"time"
)

type Config struct {
	ContextName string
	Command     string
	Args        []string

	Forge       string
	Commit      string
	PR          string
	URL         string
	PendingDesc string
	SuccessDesc string
	FailureDesc string
	Timeout     time.Duration
	Silent      bool
}
