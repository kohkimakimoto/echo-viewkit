package subprocess

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunWithContext(t *testing.T) {
	var stdout bytes.Buffer
	err := RunWithContext(context.Background(), &Subprocess{
		Command: "echo",
		Args:    []string{"hello world"},
		Stdout:  &stdout,
	})
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "hello world") {
		t.Errorf("Expected 'hello world', got: %s", stdout.String())
	}
}

func TestRunWithContext_Timeout(t *testing.T) {
	var stdout bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := RunWithContext(ctx, &Subprocess{
		Command: "sleep",
		Args:    []string{"10"},
		Stdout:  &stdout,
	})
	if err == nil {
		t.Fatal("Expected timeout error")
	}
	t.Logf("Expected timeout error, got: %v", err)
}
