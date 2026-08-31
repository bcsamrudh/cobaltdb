package main

import (
	"strings"
	"testing"

	"github.com/bcsamrudh/cobaltdb/internal/protocol"
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

func TestBuildKVRequest(t *testing.T) {
	tests := []struct {
		args    []string
		want    string
		wantErr bool
	}{
		{[]string{"put", "name", "samrudh"}, "SET name samrudh", false},
		{[]string{"put", "greeting", "hello", "world"}, "SET greeting hello world", false},
		{[]string{"get", "name"}, "GET name", false},
		{[]string{"delete", "name"}, "DELETE name", false},
		{[]string{"exists", "name"}, "EXISTS name", false},
		{nil, "", true},
		{[]string{"put", "name"}, "", true},
		{[]string{"get", "name", "extra"}, "", true},
		{[]string{"unknown", "name"}, "", true},
	}
	for _, test := range tests {
		got, err := buildKVRequest(test.args)
		if (err != nil) != test.wantErr {
			t.Errorf("buildKVRequest(%v) error = %v, wantErr %v", test.args, err, test.wantErr)
		}
		if !test.wantErr && got != test.want {
			t.Errorf("buildKVRequest(%v) = %q, want %q", test.args, got, test.want)
		}
	}
}

func TestDisplayResponse(t *testing.T) {
	tests := []struct {
		response protocol.Response
		want     string
		wantErr  bool
	}{
		{protocol.OK, "OK", false},
		{protocol.Deleted, "OK", false},
		{protocol.NotFound, "(nil)", false},
		{protocol.ExistsYes, "true", false},
		{protocol.ExistsNo, "false", false},
		{protocol.Value("samrudh"), "samrudh", false},
		{protocol.Response("ERROR bad command"), "", true},
		{protocol.Response("INVALID"), "", true},
	}
	for _, test := range tests {
		got, err := displayResponse(test.response)
		if (err != nil) != test.wantErr {
			t.Errorf("displayResponse(%q) error = %v, wantErr %v", test.response, err, test.wantErr)
		}
		if got != test.want {
			t.Errorf("displayResponse(%q) = %q, want %q", test.response, got, test.want)
		}
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	if err := run([]string{"unknown"}); err == nil {
		t.Fatal("run(unknown) returned nil error")
	}
}
