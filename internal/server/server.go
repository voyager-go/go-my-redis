package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-my-redis/internal/config"
	"go-my-redis/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	config *config.Config
	router *gin.Engine
	rdb    *redis.Client
}

func NewServer(config *config.Config, rdb *redis.Client) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	return &Server{
		config: config,
		router: router,
		rdb:    rdb,
	}
}

func (s *Server) setupRoutes() {
	// 静态文件服务
	dist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		panic(err)
	}
	s.router.StaticFS("/", http.FS(dist))

	// API 路由
	api := s.router.Group("/api")
	{
		api.POST("/connect", s.handleConnect)
		api.POST("/disconnect", s.handleDisconnect)
		api.POST("/execute", s.handleExecute)
		api.GET("/keys", s.handleKeys)
		api.GET("/key/:key", s.handleGetKey)
	}
}

func (s *Server) Start() error {
	s.setupRoutes()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.router,
	}

	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			fmt.Printf("Server forced to shutdown: %v\n", err)
		}
	}()

	fmt.Printf("Server is running on port %d\n", s.config.Port)
	return srv.ListenAndServe()
}
