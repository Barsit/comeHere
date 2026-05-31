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
	if cfg.CADays != 3650 {
		t.Errorf("expected 3650, got %d", cfg.CADays)
	}
	if cfg.Rules == nil {
		t.Error("Rules should not be nil")
	}
	if len(cfg.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(cfg.Rules))
	}
}

func TestSaveAndLoad(t *testing.T) {
	origHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", t.TempDir())
	defer os.Setenv("USERPROFILE", origHome)

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
	if loaded.Rules[0].Target != "localhost:8080" {
		t.Errorf("expected localhost:8080, got %s", loaded.Rules[0].Target)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	// On Windows, os.UserHomeDir() uses USERPROFILE, not HOME
	origHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", t.TempDir())
	defer os.Setenv("USERPROFILE", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("config should not be nil")
	}
}

func TestLoad_CorruptedFile(t *testing.T) {
	origHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", t.TempDir())
	defer os.Setenv("USERPROFILE", origHome)

	path, _ := ConfigPath()
	os.WriteFile(path, []byte("not-json"), 0644)

	cfg, err := Load()
	if err == nil {
		t.Error("expected error for corrupted file")
	}
	if cfg == nil {
		t.Fatal("config should fallback to default even on error")
	}
}

func TestSaveAndLoad_MultipleRules(t *testing.T) {
	origHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", t.TempDir())
	defer os.Setenv("USERPROFILE", origHome)

	cfg := DefaultConfig()
	cfg.ListenPort = 8443
	cfg.AdminPort = 9000
	cfg.CADays = 180

	cfg.Rules = append(cfg.Rules,
		HijackRule{ID: "r1", Source: "api.openai.com", Target: "localhost:3000", Enabled: true},
		HijackRule{ID: "r2", Source: "api.anthropic.com", Target: "api.deepseek.com:443", TargetTLS: true},
	)

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.ListenPort != 8443 {
		t.Errorf("expected 8443, got %d", loaded.ListenPort)
	}
	if len(loaded.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(loaded.Rules))
	}
	if loaded.Rules[1].TargetTLS != true {
		t.Error("TargetTLS should be preserved")
	}
}

func TestLoad_ZeroValueDefaults(t *testing.T) {
	origHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", t.TempDir())
	defer os.Setenv("USERPROFILE", origHome)

	cfg := &Config{}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.ListenPort != 443 {
		t.Errorf("expected default 443, got %d", loaded.ListenPort)
	}
	if loaded.AdminPort != 8848 {
		t.Errorf("expected default 8848, got %d", loaded.AdminPort)
	}
	if loaded.CADays != 3650 {
		t.Errorf("expected default 3650, got %d", loaded.CADays)
	}
	if loaded.Rules == nil {
		t.Error("Rules should not be nil")
	}
}

func TestConfigDir_CreatesDirectory(t *testing.T) {
	origHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", t.TempDir())
	defer os.Setenv("USERPROFILE", origHome)

	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir == "" {
		t.Error("ConfigDir should return non-empty path")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("ConfigDir should create the directory")
	}
}
