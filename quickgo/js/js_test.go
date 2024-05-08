package js_test

import (
	"fmt"
	"testing"

	"github.com/Nigel2392/quickgo/v2/quickgo/js"
)

func NewScript(funcName string, retValue any) string {
	return fmt.Sprintf(`
		function %s() {
			return %v;
		}
	`, funcName, retValue)
}

func TestRunScriptOK(t *testing.T) {
	var (
		funcName = "testMainFunc"
		retValue = 0
		script   = NewScript(funcName, retValue)
	)

	if err := js.NewScript(funcName).Run(script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunScriptErrorStatusCode(t *testing.T) {
	var (
		funcName = "testMainFunc"
		retValue = 1
		script   = NewScript(funcName, retValue)
	)

	if err := js.NewScript(funcName).Run(script); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
