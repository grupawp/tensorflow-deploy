package rz

// LogFormatter can be used to log to another format than JSON
type LogFormatter func(ev *Event) ([]byte, error)
