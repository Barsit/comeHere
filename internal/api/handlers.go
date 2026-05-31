package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/db/comehere/internal/config"
)

func (s *Server) listRules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"rules": s.Config.Rules})
}

func (s *Server) createRule(c *gin.Context) {
	var rule config.HijackRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.ID = fmt.Sprintf("rule-%d", time.Now().UnixNano())
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	rule.Enabled = false

	for _, r := range s.Config.Rules {
		if r.Source == rule.Source {
			c.JSON(http.StatusConflict, gin.H{"error": "domain already exists"})
			return
		}
	}

	s.Config.Rules = append(s.Config.Rules, rule)
	config.Save(s.Config)
	c.JSON(http.StatusCreated, rule)
}

func (s *Server) updateRule(c *gin.Context) {
	id := c.Param("id")
	var updated config.HijackRule
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, r := range s.Config.Rules {
		if r.ID == id {
			if r.Enabled {
				c.JSON(http.StatusBadRequest, gin.H{"error": "disable rule first before editing"})
				return
			}
			updated.ID = id
			updated.Enabled = false
			updated.CreatedAt = r.CreatedAt
			updated.UpdatedAt = time.Now()
			s.Config.Rules[i] = updated
			config.Save(s.Config)
			c.JSON(http.StatusOK, updated)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
}

func (s *Server) deleteRule(c *gin.Context) {
	id := c.Param("id")
	for i, r := range s.Config.Rules {
		if r.ID == id {
			if r.Enabled {
				c.JSON(http.StatusBadRequest, gin.H{"error": "disable rule first before deleting"})
				return
			}
			s.Config.Rules = append(s.Config.Rules[:i], s.Config.Rules[i+1:]...)
			config.Save(s.Config)
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
}

func (s *Server) enableRule(c *gin.Context) {
	id := c.Param("id")
	for i, r := range s.Config.Rules {
		if r.ID == id {
			if r.Enabled {
				c.JSON(http.StatusOK, r)
				return
			}

			if err := s.HostsMgr.Add(r.Source); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "add hosts: " + err.Error()})
				return
			}

			s.Proxy.AddRule(&s.Config.Rules[i])

			s.Config.Rules[i].Enabled = true
			s.Config.Rules[i].UpdatedAt = time.Now()
			config.Save(s.Config)

			c.JSON(http.StatusOK, s.Config.Rules[i])
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
}

func (s *Server) disableRule(c *gin.Context) {
	id := c.Param("id")
	for i, r := range s.Config.Rules {
		if r.ID == id {
			if !r.Enabled {
				c.JSON(http.StatusOK, r)
				return
			}

			if err := s.HostsMgr.Remove(r.Source); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "remove hosts: " + err.Error()})
				return
			}

			s.Proxy.RemoveRule(r.Source)

			s.Config.Rules[i].Enabled = false
			s.Config.Rules[i].UpdatedAt = time.Now()
			config.Save(s.Config)

			c.JSON(http.StatusOK, s.Config.Rules[i])
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
}

func (s *Server) getStatus(c *gin.Context) {
	enabledCount := 0
	for _, r := range s.Config.Rules {
		if r.Enabled {
			enabledCount++
		}
	}
	orphaned, _ := s.HostsMgr.ListManaged()
	c.JSON(http.StatusOK, gin.H{
		"rules_total":    len(s.Config.Rules),
		"rules_enabled":  enabledCount,
		"listen_port":    s.Config.ListenPort,
		"admin_port":     s.Config.AdminPort,
		"orphaned_hosts": orphaned,
	})
}

func (s *Server) cleanupHosts(c *gin.Context) {
	if err := s.HostsMgr.Cleanup(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "hosts cleaned"})
}

func (s *Server) certStatus(c *gin.Context) {
	caPath := s.CertMgr.Dir + string(os.PathSeparator) + "ca.pem"
	_, err := os.Stat(caPath)
	installed := err == nil
	c.JSON(http.StatusOK, gin.H{"ca_installed": installed})
}
