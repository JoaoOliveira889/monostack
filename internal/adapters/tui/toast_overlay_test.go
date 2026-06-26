package tui

import (
	"strings"
	"testing"

	"monostack/internal/domain"
)

func TestOverlayString_Simple(t *testing.T) {
	bg := "Line 1: background text here\nLine 2: background text here"
	fg := "OVERLAY"

	// Overlay fg onto bg on row 0 at column 8
	result := overlayString(bg, fg, 0, 8)
	expected := "Line 1: OVERLAYund text here\nLine 2: background text here"
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestOverlayString_WithANSI(t *testing.T) {
	bg := "\x1b[31mRed background text\x1b[0m\n\x1b[32mGreen background text\x1b[0m"
	fg := "\x1b[34mBlue overlay\x1b[0m"

	// Overlay blue onto green background text (row 1) at column 6
	// Green background text is 21 characters long visually.
	// "Green " is 6 characters. So overlay starts after "Green ".
	result := overlayString(bg, fg, 1, 6)

	// We expect the first 6 characters to remain green: \x1b[32mGreen \x1b[0m (or similar)
	// followed by the blue overlay, then followed by the remaining green characters.
	if !strings.Contains(result, "Blue overlay") {
		t.Errorf("Expected result to contain overlay, got:\n%q", result)
	}
	if !strings.Contains(result, "\x1b[32m") {
		t.Errorf("Expected result to retain green formatting, got:\n%q", result)
	}
}

func TestOverlayString_OutofBounds(t *testing.T) {
	bg := "Short"
	fg := "Overlay"

	// Overlay starting at column 10, past background end
	result := overlayString(bg, fg, 0, 10)
	expected := "Short     Overlay"
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestCopyS3PathCmd(t *testing.T) {
	m := mkModel()
	m.activeTab = panelS3
	m.s3Focus = focusObjects
	m.buckets = []domain.S3Bucket{{Name: "test-bucket"}}
	m.selectedBucketIndex = 0
	m.objects = []domain.S3Object{{Key: "test-folder/test-file.txt"}}
	m.selectedObjectIndex = 0

	cmd := m.copyS3PathCmd()
	if cmd == nil {
		t.Fatal("expected copyS3PathCmd to return a command, got nil")
	}

	msg := cmd()
	status, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg)
	}

	expected := "S3 path copied: s3://test-bucket/test-folder/test-file.txt"
	if status.Message != expected {
		t.Errorf("expected status message %q, got %q", expected, status.Message)
	}
}
