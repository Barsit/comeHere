package proxy

import (
	"testing"

	"github.com/db/comehere/internal/config"
)

func TestAddRemoveRule(t *testing.T) {
	e := New(nil, 443)
	rule := &config.HijackRule{
		ID: "test-1", Source: "api.test.com", Target: "localhost:8080",
	}
	e.AddRule(rule)

	if !e.HasRule("api.test.com") {
		t.Error("rule should exist after AddRule")
	}

	e.RemoveRule("api.test.com")
	if e.HasRule("api.test.com") {
		t.Error("rule should not exist after RemoveRule")
	}
}

func TestRuleConcurrency(t *testing.T) {
	e := New(nil, 443)
	rule := &config.HijackRule{
		ID: "test-1", Source: "api.test.com", Target: "localhost:8080",
	}

	for i := 0; i < 100; i++ {
		go e.AddRule(rule)
	}
	for i := 0; i < 100; i++ {
		go e.HasRule("api.test.com")
	}
}
