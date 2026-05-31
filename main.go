package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/db/comehere/internal/api"
	"github.com/db/comehere/internal/certmgr"
	"github.com/db/comehere/internal/config"
	"github.com/db/comehere/internal/elevate"
	"github.com/db/comehere/internal/hostsmgr"
	"github.com/db/comehere/internal/proxy"
)

func init() {
	if !elevate.IsAdmin() {
		fmt.Println("需要管理员权限，正在请求提权...")
		if err := elevate.RestartElevated(); err != nil {
			fmt.Printf("提权失败: %v\n", err)
			fmt.Println("请手动以管理员身份运行此程序")
		}
		os.Exit(0)
	}
}

//go:embed web/dist/*
var webFS embed.FS

func main() {
	fmt.Println("ComeHere ⚡ - API流量劫持工具")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Printf("load config: %v", err)
	}

	cm := certmgr.New(configDir())
	if err := cm.EnsureCA(cfg.CADays); err != nil {
		log.Fatalf("CA init: %v", err)
	}
	if err := cm.InstallToSystem(); err != nil {
		log.Printf("install CA (非致命): %v", err)
	}

	hm := hostsmgr.New()
	detectOrphanedHosts(hm)

	pe := proxy.New(cm, cfg.ListenPort)
	for i := range cfg.Rules {
		if cfg.Rules[i].Enabled {
			pe.AddRule(&cfg.Rules[i])
			hm.Add(cfg.Rules[i].Source)
		}
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	apiSrv := api.New(cfg, hm, cm, pe)
	apiSrv.RegisterRoutes(r)
	setupFrontend(r)

	go func() {
		if err := pe.Start(); err != nil {
			log.Fatalf("proxy start: %v", err)
		}
	}()

	go func() {
		addr := fmt.Sprintf(":%d", cfg.AdminPort)
		log.Printf("🌐 管理界面: http://localhost%s", addr)
		log.Printf("   请用管理员权限启动以确保 hosts 和证书操作正常")
		if err := r.Run(addr); err != nil {
			log.Fatalf("api start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在停止...")
	pe.Stop()
	log.Println("已停止")
	log.Println("⚠️ 请记得在 WebUI 中暂停所有规则以清理 hosts")
}

func configDir() string {
	dir, err := config.ConfigDir()
	if err != nil {
		return os.Getenv("USERPROFILE") + "/.comehere"
	}
	return dir
}

func setupFrontend(r *gin.Engine) {
	subFS, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		log.Printf("前端未构建（开发模式）: %v", err)
		return
	}
	r.Use(func(c *gin.Context) {
		if len(c.Request.URL.Path) > 3 && c.Request.URL.Path[:4] == "/api" {
			c.Next()
			return
		}
		c.FileFromFS(c.Request.URL.Path, http.FS(subFS))
	})
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) > 3 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.FileFromFS("index.html", http.FS(subFS))
	})
}

func detectOrphanedHosts(hm *hostsmgr.Manager) {
	managed, err := hm.ListManaged()
	if err != nil {
		log.Printf("检测 hosts 失败: %v", err)
		return
	}
	if len(managed) > 0 {
		log.Printf("⚠️ 检测到 hosts 残留: %v", managed)
	}
}
