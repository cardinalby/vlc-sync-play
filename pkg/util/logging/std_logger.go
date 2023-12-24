package logging

import (
	"io"
	"log"
)

type stdLogger struct {
	writer io.Writer
	logger *log.Logger
}

func NewLogger(writer io.Writer) Logger {
	return &stdLogger{
		writer: writer,
		logger: log.New(writer, "", log.Lmsgprefix),
	}
}

func (l *stdLogger) Info(format string, v ...any) {
	l.logger.Printf(format, v...)
}

func (l *stdLogger) Err(format string, v ...any) {
	l.logger.Printf(format, v...)
}

func (l *stdLogger) WithPrefix(prefix string) Logger {
	return &stdLogger{
		logger: log.New(l.writer, prefix+" ", log.Lmsgprefix),
	}
}
