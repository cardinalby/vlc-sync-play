package logging

type Logger interface {
	Info(format string, v ...any)
	Err(format string, v ...any)
	WithPrefix(prefix string) Logger
}
