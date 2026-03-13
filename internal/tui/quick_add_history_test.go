package tui

import "testing"

func TestQuickAddHistory_BasicNavigation(t *testing.T) {
	t.Parallel()

	var h quickAddHistory
	h.Add("first task")
	h.Add("second task")

	if got := h.Prev(""); got != "second task" {
		t.Fatalf("expected newest entry first, got %q", got)
	}
	if got := h.Prev(""); got != "first task" {
		t.Fatalf("expected previous entry, got %q", got)
	}
	if got := h.Next(""); got != "second task" {
		t.Fatalf("expected forward navigation, got %q", got)
	}
	if got := h.Next(""); got != "" {
		t.Fatalf("expected draft restore on leaving history, got %q", got)
	}
}

func TestQuickAddHistory_DedupConsecutive(t *testing.T) {
	t.Parallel()

	var h quickAddHistory
	h.Add("same")
	h.Add("same")
	if len(h.items) != 1 {
		t.Fatalf("expected consecutive duplicate suppression, got %d", len(h.items))
	}
}
