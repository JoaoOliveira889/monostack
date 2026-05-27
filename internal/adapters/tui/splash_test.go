package tui

import (
	"strings"
	"testing"
)

func TestSplashStaysVisibleUntilFrameLimit(t *testing.T) {
	m := mkModel()
	m.showSplash = true

	for i := 0; i < splashFrameLimit-1; i++ {
		result, _ := m.Update(splashTickMsg{})
		updated, ok := result.(Model)
		if !ok {
			t.Fatalf("expected Model, got %T", result)
		}
		m = updated
		if !m.showSplash {
			t.Fatalf("expected splash to remain visible before frame limit at frame %d", i+1)
		}
	}

	result, _ := m.Update(splashTickMsg{})
	updated, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	m = updated
	if m.showSplash {
		t.Fatal("expected splash to close at frame limit")
	}
}

func TestSplashUpdateReturnsFollowupCmdBeforeLimit(t *testing.T) {
	m := mkModel()
	m.showSplash = true
	m.splashFrame = splashFrameLimit - 2

	_, cmd := m.Update(splashTickMsg{})
	if cmd == nil {
		t.Fatal("expected followup splash tick command before limit")
	}
}

func TestSplashUpdateStopsAtLimit(t *testing.T) {
	m := mkModel()
	m.showSplash = true
	m.splashFrame = splashFrameLimit - 1

	result, cmd := m.Update(splashTickMsg{})
	updated, ok := result.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", result)
	}
	m = updated
	if cmd != nil {
		t.Fatal("expected nil command when splash reaches frame limit")
	}
	if m.showSplash {
		t.Fatal("expected splash to be hidden after reaching frame limit")
	}
}

func TestRenderSplashContainsUnicodeWordmark(t *testing.T) {
	m := mkModel()
	m.width = 100
	m.height = 40

	splash := m.renderSplash()
	if !strings.Contains(splash, "MonoStack") {
		t.Fatalf("expected unicode splash wordmark, got %q", splash)
	}
}
