package command_test

import (
	"testing"

	"github.com/Nigel2392/quickgo/v2/quickgo/command"
)

var env = map[string]any{
	"key":   "value",
	"key2":  "value2",
	"int":   1,
	"bool":  true,
	"float": 1.1,
	"json": map[string]any{
		"key": "value",
	},
}

func TestExpandArgs(t *testing.T) {
	var (
		step = &command.Step{
			Command: "echo",
			Args:    []string{"$key", "$key2", "$int", "$bool", "$float", "$json"},
		}
	)

	var args, _ = step.ParseArgs(env)

	if args[0] != "value" {
		t.Errorf("expected value, got %s", step.Args[0])
	}

	if args[1] != "value2" {
		t.Errorf("expected value2, got %s", step.Args[1])
	}

	if args[2] != "1" {
		t.Errorf("expected 1, got %s", step.Args[2])
	}

	if args[3] != "true" {
		t.Errorf("expected true, got %s", step.Args[3])
	}

	if args[4] != "1.1" {
		t.Errorf("expected 1.1, got %s", step.Args[4])
	}

	if args[5] != "{\"key\":\"value\"}" {
		t.Errorf("expected {\"key\":\"value\"}, got %s", step.Args[5])
	}
}
