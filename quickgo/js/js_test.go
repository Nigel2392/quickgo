package js_test

import (
	"fmt"
	"testing"

	"github.com/Nigel2392/quickgo/v2/quickgo/js"
)

func NewScript(funcName string, retValue any, retMessage string) string {
	return fmt.Sprintf(`
		function %s() {
			return %v, "%s";
		}
	`, funcName, retValue, retMessage)
}

func TestRunScriptOK(t *testing.T) {
	var (
		funcName      = "testMainFunc"
		retValue      = 0
		returnMessage = "success"
		script        = NewScript(funcName, retValue, returnMessage)
	)

	var cmd = js.NewScript(funcName)

	if d, err := cmd.Run(script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if d.Importance != retValue {
		t.Fatalf("expected exit code %d, got %d", retValue, d.Importance)
	} else if d.Message != returnMessage {
		t.Fatalf("expected message %s, got %s", returnMessage, d.Message)
	}

}

func TestRunScriptErrorStatusCode(t *testing.T) {
	var (
		funcName   = "testMainFunc"
		retValue   = 1
		retMessage = "error message"
		script     = NewScript(funcName, retValue, retMessage)
	)

	var cmd = js.NewScript(funcName)

	if d, err := cmd.Run(script); err == nil {
		t.Fatalf("expected error, got nil")
	} else if d.Importance != retValue {
		t.Fatalf("expected exit code %d, got %d", retValue, d.Importance)
	} else if d.Message != retMessage {
		t.Fatalf("expected message %s, got %s", retMessage, d.Message)
	}
}
