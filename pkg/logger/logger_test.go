package logger_test

import (
	"bytes"
	"testing"

	"github.com/rerost/llm-modeling-evaluation/pkg/logger"
)

func TestShowImmediate(t *testing.T) {
	l := logger.NewLogger()
	w := &bytes.Buffer{}

	logger.ShowImmediate(l, w)
	l.Write("Test log message")

	expected := "Test log message\n"
	if w.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, w.String())
	}

	l.Close()
}

func TestShowAfterClose(t *testing.T) {
	l := logger.NewLogger()
	w := &bytes.Buffer{}

	logger.ShowAfterClose(l, w)
	l.Write("Test log message")

	expected := ""
	if w.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, w.String())
	}

	l.Close()

	expected = "Test log message\n"
	if w.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, w.String())
	}
}
