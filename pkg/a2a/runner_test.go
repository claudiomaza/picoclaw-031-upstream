package a2a

import (
	"context"
	"testing"
)

func TestExecuteRejectsEmptyMessage(t *testing.T) {
	r := &Runner{}
	if _, err := r.Execute(context.Background(), TurnRequest{}); err == nil {
		t.Fatal("expected empty message error")
	}
}
