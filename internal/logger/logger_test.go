package logger

import "testing"

func TestNewLoggerLevelsAndFormat(t *testing.T) {
	log := New("debug", "json")
	if log == nil || log.Logger == nil {
		t.Fatalf("expected logger instance")
	}
	if log.GetLevel().String() != "debug" {
		t.Fatalf("expected debug level, got %s", log.GetLevel().String())
	}

	infoLog := New("unknown-level", "text")
	if infoLog.GetLevel().String() != "info" {
		t.Fatalf("expected default info level, got %s", infoLog.GetLevel().String())
	}
}

func TestWithFieldAndWithFields(t *testing.T) {
	log := New("info", "text")
	if entry := log.WithField("k", "v"); entry == nil {
		t.Fatalf("expected entry from WithField")
	}
	if entry := log.WithFields(Fields{"a": 1, "b": "x"}); entry == nil {
		t.Fatalf("expected entry from WithFields")
	}
}
