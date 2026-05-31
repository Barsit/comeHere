package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/db/comehere/internal/certmgr"
	"github.com/db/comehere/internal/config"
)

type Engine struct {
	certMgr  *certmgr.Manager
	rules    map[string]*config.HijackRule
	mu       sync.RWMutex
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	port     int
}

func New(certMgr *certmgr.Manager, port int) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		certMgr: certMgr,
		rules:   make(map[string]*config.HijackRule),
		port:    port,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (e *Engine) AddRule(rule *config.HijackRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules[rule.Source] = rule
}

func (e *Engine) RemoveRule(domain string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.rules, domain)
}

func (e *Engine) HasRule(domain string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.rules[domain]
	return ok
}

func (e *Engine) Start() error {
	tlsCfg := &tls.Config{
		GetCertificate: e.getCertificate,
	}
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", e.port), tlsCfg)
	if err != nil {
		return fmt.Errorf("listen :%d: %w", e.port, err)
	}
	e.listener = listener
	log.Printf("🔒 TLS proxy listening on :%d", e.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-e.ctx.Done():
				return nil
			default:
				log.Printf("accept error: %v", err)
				continue
			}
		}
		go e.handleConn(conn)
	}
}

func (e *Engine) Stop() error {
	e.cancel()
	if e.listener != nil {
		return e.listener.Close()
	}
	return nil
}

func (e *Engine) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	e.mu.RLock()
	_, ok := e.rules[hello.ServerName]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown domain: %s", hello.ServerName)
	}

	cert, err := e.certMgr.IssueCert(hello.ServerName)
	if err != nil {
		return nil, fmt.Errorf("issue cert for %s: %w", hello.ServerName, err)
	}
	return cert, nil
}
