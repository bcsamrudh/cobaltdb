package protocol

import (
	"testing"

	"github.com/bcsamrudh/cobaltdb/internal/store"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    Command
		wantErr bool
	}{
		{"set", "SET name samrudh", Command{Set, "name", "samrudh"}, false},
		{"case insensitive", "get name", Command{Get, "name", ""}, false},
		{"value with spaces", "SET greeting hello world", Command{Set, "greeting", "hello world"}, false},
		{"delete", "DELETE name", Command{Delete, "name", ""}, false},
		{"exists", "EXISTS name", Command{Exists, "name", ""}, false},
		{"empty", "", Command{}, true},
		{"set missing value", "SET name", Command{}, true},
		{"get extra argument", "GET name extra", Command{}, true},
		{"unknown", "RENAME old new", Command{}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.line)
			if (err != nil) != test.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", test.line, err, test.wantErr)
			}
			if !test.wantErr && got != test.want {
				t.Fatalf("Parse(%q) = %#v, want %#v", test.line, got, test.want)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	database := store.New()
	tests := []struct {
		line string
		want Response
	}{
		{"GET name", NotFound},
		{"SET name samrudh", OK},
		{"GET name", Value("samrudh")},
		{"EXISTS name", ExistsYes},
		{"DELETE name", Deleted},
		{"GET name", NotFound},
		{"EXISTS name", ExistsNo},
		{"DELETE name", NotFound},
		{"UNKNOWN", Response("ERROR unknown command: UNKNOWN")},
	}

	for _, test := range tests {
		if got := Execute(database, test.line); got != test.want {
			t.Errorf("Execute(%q) = %q, want %q", test.line, got, test.want)
		}
	}
}
