package certmgr

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureCA(t *testing.T) {
	dir := t.TempDir()
	m := New(filepath.Dir(dir))
	m.Dir = filepath.Join(dir, "certs")

	if err := m.EnsureCA(365); err != nil {
		t.Fatal(err)
	}

	caPath := filepath.Join(m.Dir, "ca.pem")
	if _, err := os.Stat(caPath); os.IsNotExist(err) {
		t.Error("CA cert should exist")
	}
	keyPath := filepath.Join(m.Dir, "ca-key.pem")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("CA key should exist")
	}
}

func TestIssueCert(t *testing.T) {
	dir := t.TempDir()
	m := New(filepath.Dir(dir))
	m.Dir = filepath.Join(dir, "certs")

	if err := m.EnsureCA(365); err != nil {
		t.Fatal(err)
	}

	cert, err := m.IssueCert("api.test.com")
	if err != nil {
		t.Fatal(err)
	}
	if cert == nil {
		t.Fatal("cert should not be nil")
	}

	parsed, _ := x509.ParseCertificate(cert.Certificate[0])
	if len(parsed.DNSNames) == 0 || parsed.DNSNames[0] != "api.test.com" {
		t.Errorf("expected DNSNames api.test.com, got %v", parsed.DNSNames)
	}
}

func TestIssueCertCache(t *testing.T) {
	dir := t.TempDir()
	m := New(filepath.Dir(dir))
	m.Dir = filepath.Join(dir, "certs")

	m.EnsureCA(365)
	cert1, _ := m.IssueCert("api.test.com")
	cert2, _ := m.IssueCert("api.test.com")

	if len(cert1.Certificate[0]) != len(cert2.Certificate[0]) {
		t.Error("cached cert should match")
	}
}
