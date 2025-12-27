package sentry

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

type Monitor struct {
	Slug string
	Hub  *sentry.Hub
	ID   *sentry.EventID
}

func NewMonitor(dsn, slug string) (*Monitor, error) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		return nil, err
	}
	return &Monitor{
		Slug: slug,
		Hub:  sentry.CurrentHub(),
	}, nil
}

func (m *Monitor) Start() {
	id := m.Hub.CaptureCheckIn(&sentry.CheckIn{
		MonitorSlug: m.Slug,
		Status:      sentry.CheckInStatusInProgress,
	}, nil)
	m.ID = id
}

func (m *Monitor) Finish(success bool) {
	status := sentry.CheckInStatusOK
	if !success {
		status = sentry.CheckInStatusError
	}

	if m.ID == nil {
		return
	}

	m.Hub.CaptureCheckIn(&sentry.CheckIn{
		ID:          *m.ID,
		MonitorSlug: m.Slug,
		Status:      status,
		Duration:    0, // Sentry calculates duration automatically if we use the ID
	}, nil)

	// Flush to ensure events are sent before the program exits
	m.Hub.Flush(2 * time.Second)
}

// EnvDSN returns the Sentry DSN from the environment
func EnvDSN() string {
	return os.Getenv("SENTRY_DSN")
}
