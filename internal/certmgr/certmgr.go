package certmgr

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

type Manager struct {
	Dir       string
	caCert    *x509.Certificate
	caKey     interface{}
	caTLSCert *tls.Certificate
}

// execCommand 可被测试 mock
var execCommand = func(name string, arg ...string) error {
	return fmt.Errorf("not implemented: %s %v", name, arg)
}

func New(configDir string) *Manager {
	return &Manager{Dir: filepath.Join(configDir, "certs")}
}

func (m *Manager) EnsureCA(days int) error {
	if err := os.MkdirAll(m.Dir, 0755); err != nil {
		return err
	}
	caPath := filepath.Join(m.Dir, "ca.pem")
	keyPath := filepath.Join(m.Dir, "ca-key.pem")

	if _, err := os.Stat(caPath); err == nil {
		return m.loadCA(caPath, keyPath)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate CA key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "ComeHere Root CA",
			Organization: []string{"ComeHere"},
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().AddDate(0, 0, days),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("create CA cert: %w", err)
	}

	certFile, err := os.Create(caPath)
	if err != nil {
		return err
	}
	defer certFile.Close()
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyFile, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyFile.Close()
	pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return m.loadCA(caPath, keyPath)
}

func (m *Manager) loadCA(certPath, keyPath string) error {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	tlsCert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return err
	}
	m.caTLSCert = &tlsCert

	m.caCert, err = x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return err
	}
	m.caKey = tlsCert.PrivateKey
	return nil
}

func (m *Manager) IssueCert(domain string) (*tls.Certificate, error) {
	if m.caCert == nil || m.caKey == nil {
		return nil, fmt.Errorf("CA not initialized")
	}

	certPath := filepath.Join(m.Dir, domain+".pem")
	keyPath := filepath.Join(m.Dir, domain+"-key.pem")

	if _, err := os.Stat(certPath); err == nil {
		if cert, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil {
			return &cert, nil
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: domain,
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().AddDate(0, 0, 365),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		DNSNames: []string{domain},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, m.caCert, &key.PublicKey, m.caKey)
	if err != nil {
		return nil, fmt.Errorf("create cert: %w", err)
	}

	certFile, _ := os.Create(certPath)
	defer certFile.Close()
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyFile, _ := os.Create(keyPath)
	defer keyFile.Close()
	pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return &tls.Certificate{
		Certificate: [][]byte{certDER, m.caTLSCert.Certificate[0]},
		PrivateKey:  key,
	}, nil
}

func (m *Manager) InstallToSystem() error {
	certPath := filepath.Join(m.Dir, "ca.pem")
	return execCommand("certutil", "-addstore", "-f", "Root", certPath)
}
