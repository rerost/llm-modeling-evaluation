package logger

import (
	"io"
)

type Logger struct {
	log    chan string
	closed chan struct{}
}

func (l *Logger) Write(log string) {
	l.log <- log
}

func (l *Logger) Close() {
	close(l.log)
	<-l.closed
}

func (l *Logger) watch() <-chan string {
	return l.log
}

func NewLogger() *Logger {
	return &Logger{
		log:    make(chan string),
		closed: make(chan struct{}),
	}
}

func ShowImmediate(l *Logger, writer io.Writer) {
	go func() {
		for log := range l.watch() {
			writer.Write([]byte(log + "\n"))
		}

		close(l.closed)
	}()
}

func ShowAfterClose(l *Logger, writer io.Writer) {
	go func() {
		logs := []string{}
		for log := range l.watch() {
			logs = append(logs, log)
		}

		for _, log := range logs {
			writer.Write([]byte(log + "\n"))
		}

		close(l.closed)
	}()
}
