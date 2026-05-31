package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ListenPort != 443 {
		t.Errorf("expected 443, got %d", cfg.ListenPort)
	}
	if cfg.AdminPort != 8848 {
		t.Errorf("expected 8848, got %d", cfg.AdminPort)
	}
	if cfg.Rules == nil {
		t.Error("Rules should not be nil")
	}
}

func TestSaveAndLoad(t *testing.T) {
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", origHome)

	cfg := DefaultConfig()
	cfg.Rules = append(cfg.Rules, HijackRule{
		ID: "test-1", Source: "api.example.com", Target: "localhost:8080",
	})
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(loaded.Rules))
	}
	if loaded.Rules[0].Source != "api.example.com" {
		t.Errorf("expected api.example.com, got %s", loaded.Rules[0].Source)
	}
}
