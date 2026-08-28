package main

import (
	"strings"
	"testing"
)

func TestParseServerOptions(t *testing.T) {
	options, err := parseServerOptions([]string{"-dev", "-addr", "127.0.0.1:7000"})
	if err != nil {
		t.Fatal(err)
	}
	if !options.dev {
		t.Fatal("dev = false, want true")
	}
	if options.address != "127.0.0.1:7000" {
		t.Fatalf("address = %q, want 127.0.0.1:7000", options.address)
	}
}

func TestParseServerOptionsDefaults(t *testing.T) {
	options, err := parseServerOptions(nil)
	if err != nil {
		t.Fatal(err)
	}
	if options.dev {
		t.Fatal("dev = true, want false")
	}
	if options.address != "127.0.0.1:6380" {
		t.Fatalf("address = %q, want 127.0.0.1:6380", options.address)
	}
}

func TestRunRequiresCommand(t *testing.T) {
	err := run(nil)
	if err == nil || !strings.Contains(err.Error(), "command is required") {
		t.Fatalf("run(nil) error = %v", err)
	}
}

func TestRunRejectsUnavailableCommands(t *testing.T) {
	tests := [][]string{
		{"kv", "get", "name"},
		{"unknown"},
	}
	for _, args := range tests {
		if err := run(args); err == nil {
			t.Errorf("run(%v) returned nil error", args)
		}
	}
}
