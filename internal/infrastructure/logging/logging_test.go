package logging

import (
	"log/slog"
	"testing"
)

func TestParseLevel_Debug(t *testing.T) {
	lvl, err := ParseLevel("debug")
	if err != nil {
		t.Fatalf("ParseLevel(\"debug\") error: %v", err)
	}
	if lvl != slog.LevelDebug {
		t.Errorf("got %v, want %v", lvl, slog.LevelDebug)
	}
}

func TestParseLevel_Info(t *testing.T) {
	lvl, err := ParseLevel("info")
	if err != nil {
		t.Fatalf("ParseLevel(\"info\") error: %v", err)
	}
	if lvl != slog.LevelInfo {
		t.Errorf("got %v, want %v", lvl, slog.LevelInfo)
	}
}

func TestParseLevel_Warn(t *testing.T) {
	lvl, err := ParseLevel("warn")
	if err != nil {
		t.Fatalf("ParseLevel(\"warn\") error: %v", err)
	}
	if lvl != slog.LevelWarn {
		t.Errorf("got %v, want %v", lvl, slog.LevelWarn)
	}
}

func TestParseLevel_Error(t *testing.T) {
	lvl, err := ParseLevel("error")
	if err != nil {
		t.Fatalf("ParseLevel(\"error\") error: %v", err)
	}
	if lvl != slog.LevelError {
		t.Errorf("got %v, want %v", lvl, slog.LevelError)
	}
}

func TestParseLevel_CaseInsensitive(t *testing.T) {
	lvl, err := ParseLevel("DEBUG")
	if err != nil {
		t.Fatalf("ParseLevel(\"DEBUG\") error: %v", err)
	}
	if lvl != slog.LevelDebug {
		t.Errorf("got %v, want %v", lvl, slog.LevelDebug)
	}
}

func TestParseLevel_Invalid(t *testing.T) {
	_, err := ParseLevel("invalid")
	if err == nil {
		t.Fatal("ParseLevel(\"invalid\") should return error")
	}
}

func TestSetup_ValidLevel(t *testing.T) {
	if err := Setup("debug"); err != nil {
		t.Fatalf("Setup(\"debug\") error: %v", err)
	}
}

func TestSetup_InvalidLevel(t *testing.T) {
	if err := Setup("invalid"); err == nil {
		t.Fatal("Setup(\"invalid\") should return error")
	}
}
