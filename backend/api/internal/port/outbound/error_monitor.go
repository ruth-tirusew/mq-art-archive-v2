package outbound

type ErrorMonitor interface {
	Capture(err error, fields map[string]string)
}
