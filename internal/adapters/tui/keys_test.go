package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestKeyMap_Bindings(t *testing.T) {
	tests := []struct {
		name string
		b    key.Binding
	}{
		{"TabS3", keys.TabS3},
		{"TabSQS", keys.TabSQS},
		{"TabSNS", keys.TabSNS},
		{"TabConfig", keys.TabConfig},
		{"Up", keys.Up},
		{"Down", keys.Down},
		{"Left", keys.Left},
		{"Right", keys.Right},
		{"Enter", keys.Enter},
		{"Esc", keys.Esc},
		{"Quit", keys.Quit},
		{"S3Presign", keys.S3Presign},
		{"S3Delete", keys.S3Delete},
		{"S3Download", keys.S3Download},
		{"S3Folder", keys.S3Folder},
		{"SQSPurge", keys.SQSPurge},
		{"SQSPurgeAll", keys.SQSPurgeAll},
		{"SQSView", keys.SQSView},
		{"SQSSend", keys.SQSSend},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.b.Help().Key == "" {
				t.Errorf("%s: Help Key is empty", tt.name)
			}
		})
	}
}

func TestKeyMap_UniqueKeys(t *testing.T) {
	seen := make(map[string]string)
	allBindings := []struct {
		name string
		b    key.Binding
	}{
		{"TabS3", keys.TabS3},
		{"TabSQS", keys.TabSQS},
		{"TabSNS", keys.TabSNS},
		{"TabConfig", keys.TabConfig},
		{"Up", keys.Up},
		{"Down", keys.Down},
		{"Left", keys.Left},
		{"Right", keys.Right},
		{"Enter", keys.Enter},
		{"Esc", keys.Esc},
		{"Quit", keys.Quit},
		{"S3Presign", keys.S3Presign},
		{"S3Delete", keys.S3Delete},
		{"S3Download", keys.S3Download},
		{"S3Folder", keys.S3Folder},
		{"SQSPurge", keys.SQSPurge},
		{"SQSPurgeAll", keys.SQSPurgeAll},
		{"SQSView", keys.SQSView},
		{"SQSSend", keys.SQSSend},
	}

	for _, b := range allBindings {
		for _, k := range b.b.Keys() {
			if prev, ok := seen[k]; ok {
				t.Errorf("key %q is used by both %q and %q", k, prev, b.name)
			}
			seen[k] = b.name
		}
	}
}
