package errormonitor

import "log"

type Noop struct{}

func NewNoop() *Noop                           { return &Noop{} }
func (*Noop) Capture(error, map[string]string) {}

type Log struct {
	dsn string
}

func NewLog(dsn string) *Log { return &Log{dsn: dsn} }

func (m *Log) Capture(err error, fields map[string]string) {
	// The DSN is a credential for the monitoring service; never write it to logs.
	log.Printf("error_monitor error=%q request_id=%q", err.Error(), fields["request_id"])
}
