package api

import (
	"github.com/gin-gonic/gin"
	"github.com/db/comehere/internal/certmgr"
	"github.com/db/comehere/internal/config"
	"github.com/db/comehere/internal/hostsmgr"
	"github.com/db/comehere/internal/proxy"
)

type Server struct {
	Config   *config.Config
	HostsMgr *hostsmgr.Manager
	CertMgr  *certmgr.Manager
	Proxy    *proxy.Engine
}

func New(cfg *config.Config, hosts *hostsmgr.Manager, cert *certmgr.Manager, px *proxy.Engine) *Server {
	return &Server{
		Config:   cfg,
		HostsMgr: hosts,
		CertMgr:  cert,
		Proxy:    px,
	}
}

// RegisterRoutes 注册 API 路由到 gin.Engine
func (s *Server) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/rules", s.listRules)
		api.POST("/rules", s.createRule)
		api.PUT("/rules/:id", s.updateRule)
		api.DELETE("/rules/:id", s.deleteRule)
		api.POST("/rules/:id/enable", s.enableRule)
		api.POST("/rules/:id/disable", s.disableRule)
		api.GET("/status", s.getStatus)
		api.POST("/cleanup", s.cleanupHosts)
		api.GET("/cert/status", s.certStatus)
	}
}
