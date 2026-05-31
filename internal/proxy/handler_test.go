package proxy

import (
	"testing"
)

func TestExtractHost(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"localhost:3000", "localhost"},
		{"127.0.0.1:443", "127.0.0.1"},
		{"api.deepseek.com:443", "api.deepseek.com"},
		{"192.168.1.100:8080", "192.168.1.100"},
		{"no-port", "no-port"},
		{"", ""},
		// IPv6 with port: the function finds the last colon before the port number
		{"[::1]:3000", "[::1]"},
		{"[::1]:443", "[::1]"},
	}

	for _, tt := range tests {
		result := extractHost(tt.input)
		if result != tt.expected {
			t.Errorf("extractHost(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractHostEdgeCases(t *testing.T) {
	// Empty string returns empty
	if r := extractHost(""); r != "" {
		t.Errorf("expected empty, got %q", r)
	}

	// Single colon returns colon (idx=0, not >0, returns input)
	if r := extractHost(":"); r != ":" {
		t.Errorf("expected ':', got %q", r)
	}

	// Multiple colons without port returns all but last segment
	// e.g. "::1" → last colon at idx 2 → returns ":"
	if r := extractHost("::1"); r != ":" {
		t.Errorf("expected ':', got %q", r)
	}

	// IP address without port
	if r := extractHost("192.168.1.1"); r != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got %q", r)
	}

	// Hostname with multiple colons and port
	if r := extractHost("ipv6:::1:3000"); r != "ipv6:::1" {
		t.Errorf("expected 'ipv6:::1', got %q", r)
	}
}
