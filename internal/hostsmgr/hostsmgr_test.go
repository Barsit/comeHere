package hostsmgr

import (
	"os"
	"testing"
)

func writeTestHosts(t *testing.T, content string) string {
	t.Helper()
	path := t.TempDir() + "\\hosts"
	os.WriteFile(path, []byte(content), 0644)
	return path
}

func TestAddDomain(t *testing.T) {
	content := "127.0.0.1 localhost"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	if err := m.Add("api.test.com"); err != nil {
		t.Fatal(err)
	}

	has, _ := m.HasDomain("api.test.com")
	if !has {
		t.Error("domain should exist after Add")
	}
}

func TestAddDomain_PreserveExisting(t *testing.T) {
	content := "127.0.0.1 localhost\n127.0.0.1 existing.com"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	if err := m.Add("api.test.com"); err != nil {
		t.Fatal(err)
	}

	has, _ := m.HasDomain("existing.com")
	if !has {
		t.Error("existing entries should be preserved")
	}
	has, _ = m.HasDomain("api.test.com")
	if !has {
		t.Error("new domain should exist")
	}
}

func TestRemoveDomain(t *testing.T) {
	content := "127.0.0.1 localhost\n# ComeHere\n127.0.0.1 api.test.com\n"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	if err := m.Remove("api.test.com"); err != nil {
		t.Fatal(err)
	}

	has, _ := m.HasDomain("api.test.com")
	if has {
		t.Error("domain should not exist after Remove")
	}
}

func TestRemoveDomain_NonExistent(t *testing.T) {
	content := "127.0.0.1 localhost\n# ComeHere\n127.0.0.1 api.test.com\n"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	// Removing a non-existent domain should not error
	err := m.Remove("nonexistent.com")
	if err != nil {
		t.Errorf("removing non-existent domain should not error: %v", err)
	}

	// Original managed domain should still exist
	has, _ := m.HasDomain("api.test.com")
	if !has {
		t.Error("other managed domains should still exist")
	}
}

func TestCleanup(t *testing.T) {
	content := "127.0.0.1 localhost\n# ComeHere\n127.0.0.1 api.test.com\n"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	if err := m.Cleanup(); err != nil {
		t.Fatal(err)
	}

	domains, _ := m.ListManaged()
	if len(domains) != 0 {
		t.Errorf("expected 0 managed domains, got %d", len(domains))
	}
}

func TestCleanup_MultipleEntries(t *testing.T) {
	content := "127.0.0.1 localhost\n# ComeHere\n127.0.0.1 api.test.com\n# ComeHere\n127.0.0.1 api2.com\n"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	if err := m.Cleanup(); err != nil {
		t.Fatal(err)
	}

	domains, _ := m.ListManaged()
	if len(domains) != 0 {
		t.Errorf("expected 0 domains after cleanup, got %d", len(domains))
	}
	// localhost should still exist
	contentBytes, _ := os.ReadFile(path)
	if !contains(string(contentBytes), "127.0.0.1 localhost") {
		t.Error("localhost entry should be preserved")
	}
}

func TestAddDuplicate(t *testing.T) {
	content := "127.0.0.1 localhost"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	m.Add("api.test.com")
	err := m.Add("api.test.com")
	if err != nil {
		t.Fatal(err)
	}

	domains, _ := m.ListManaged()
	if len(domains) != 1 {
		t.Errorf("expected 1 domain, got %d", len(domains))
	}
}

func TestHasDomain_NonExistent(t *testing.T) {
	content := "127.0.0.1 localhost"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	has, _ := m.HasDomain("nonexistent.com")
	if has {
		t.Error("should return false for non-existent domain")
	}
}

func TestListManaged_Empty(t *testing.T) {
	content := "127.0.0.1 localhost"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	domains, _ := m.ListManaged()
	if len(domains) != 0 {
		t.Errorf("expected 0 managed domains, got %d", len(domains))
	}
}

func TestListManaged_Multiple(t *testing.T) {
	content := "127.0.0.1 localhost\n# ComeHere\n127.0.0.1 a.com\n# ComeHere\n127.0.0.1 b.com"
	path := writeTestHosts(t, content)
	m := &Manager{Path: path}

	domains, _ := m.ListManaged()
	if len(domains) != 2 {
		t.Errorf("expected 2 managed domains, got %d", len(domains))
	}
	if domains[0] != "a.com" || domains[1] != "b.com" {
		t.Errorf("unexpected domains: %v", domains)
	}
}

func TestAdd_FileNotExist(t *testing.T) {
	m := &Manager{Path: "Z:\\nonexistent\\hosts"}
	err := m.Add("api.test.com")
	if err == nil {
		t.Error("should error when file doesn't exist")
	}
}

func TestHasDomain_FileNotExist(t *testing.T) {
	m := &Manager{Path: "Z:\\nonexistent\\hosts"}
	_, err := m.HasDomain("test.com")
	if err == nil {
		t.Error("should error when file doesn't exist")
	}
}

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
