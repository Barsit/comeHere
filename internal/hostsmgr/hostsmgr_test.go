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
