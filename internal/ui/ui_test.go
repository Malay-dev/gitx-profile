package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// captureStderr captures what's written to os.Stderr during fn execution.
func captureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestSuccess_ContainsMessage(t *testing.T) {
	output := captureStderr(func() {
		Success("Profile %q activated", "work")
	})

	if !strings.Contains(output, "Profile \"work\" activated") {
		t.Errorf("Success output missing message, got: %q", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("Success output missing checkmark, got: %q", output)
	}
}

func TestError_ContainsMessage(t *testing.T) {
	output := captureStderr(func() {
		Error("Profile %q not found", "ghost")
	})

	if !strings.Contains(output, "Profile \"ghost\" not found") {
		t.Errorf("Error output missing message, got: %q", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("Error output missing cross mark, got: %q", output)
	}
}

func TestWarn_ContainsMessage(t *testing.T) {
	output := captureStderr(func() {
		Warn("No matching profile")
	})

	if !strings.Contains(output, "No matching profile") {
		t.Errorf("Warn output missing message, got: %q", output)
	}
	if !strings.Contains(output, "⚠") {
		t.Errorf("Warn output missing warning symbol, got: %q", output)
	}
}

func TestInfo_ContainsMessage(t *testing.T) {
	output := captureStderr(func() {
		Info("Run 'git profile add' to create one")
	})

	if !strings.Contains(output, "Run 'git profile add' to create one") {
		t.Errorf("Info output missing message, got: %q", output)
	}
	if !strings.Contains(output, "ℹ") {
		t.Errorf("Info output missing info symbol, got: %q", output)
	}
}

func TestField_ContainsLabelAndValue(t *testing.T) {
	output := captureStderr(func() {
		Field("Email", "alice@acmecorp.com")
	})

	if !strings.Contains(output, "Email") {
		t.Errorf("Field output missing label, got: %q", output)
	}
	if !strings.Contains(output, "alice@acmecorp.com") {
		t.Errorf("Field output missing value, got: %q", output)
	}
}

func TestSuccess_FormatArgs(t *testing.T) {
	output := captureStderr(func() {
		Success("Switched to %s (scope: %s)", "work", "local")
	})

	if !strings.Contains(output, "Switched to work (scope: local)") {
		t.Errorf("Success format args not applied, got: %q", output)
	}
}
