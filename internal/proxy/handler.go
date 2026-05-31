package proxy

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func (e *Engine) handleConn(clientConn net.Conn) {
	defer clientConn.Close()

	clientTLS := clientConn.(*tls.Conn)
	if err := clientTLS.Handshake(); err != nil {
		log.Printf("TLS handshake error: %v", err)
		return
	}

	serverName := clientTLS.ConnectionState().ServerName
	e.mu.RLock()
	rule, ok := e.rules[serverName]
	e.mu.RUnlock()

	if !ok {
		log.Printf("no rule for domain: %s", serverName)
		return
	}

	log.Printf("→ %s → %s", rule.Source, rule.Target)

	var upstream net.Conn
	var err error

	if rule.TargetTLS {
		upstream, err = tls.DialWithDialer(
			&net.Dialer{Timeout: 10 * time.Second},
			"tcp", rule.Target,
			&tls.Config{ServerName: extractHost(rule.Target)},
		)
	} else {
		upstream, err = net.DialTimeout("tcp", rule.Target, 10*time.Second)
	}

	if err != nil {
		log.Printf("connect to %s failed: %v", rule.Target, err)
		return
	}
	defer upstream.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(upstream, clientConn)
		upstream.Close()
	}()
	go func() {
		defer wg.Done()
		io.Copy(clientConn, upstream)
		clientConn.Close()
	}()
	wg.Wait()
	log.Printf("← %s done", rule.Source)
}

func extractHost(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx > 0 {
		return addr[:idx]
	}
	return addr
}
