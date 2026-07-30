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
	log.Printf("error_monitor dsn=%q error=%q request_id=%q", m.dsn, err.Error(), fields["request_id"])
}
