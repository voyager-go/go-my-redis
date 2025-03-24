package main

import (
	"log"
	"os/exec"
	"runtime"
	"time"

	"go-my-redis/internal/handler"

	"github.com/gin-gonic/gin"
)

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("无法自动打开浏览器: %v", err)
	}
}

func main() {
	r := gin.Default()
	redisHandler := handler.NewRedisHandler()

	// API 路由
	api := r.Group("/api")
	{
		// Redis 连接管理
		api.POST("/connect", redisHandler.Connect)
		api.POST("/disconnect", redisHandler.Disconnect)

		// Redis 命令执行
		api.POST("/command", redisHandler.ExecuteCommand)

		// 键管理
		api.GET("/keys", redisHandler.GetKeys)
		api.GET("/key/:key", redisHandler.GetKey)
		api.GET("/key/:key/type", redisHandler.GetType)
		api.GET("/key/:key/ttl", redisHandler.GetTTL)

		// 列表操作
		api.GET("/list/:key", redisHandler.GetList)
		api.GET("/list/:key/length", redisHandler.GetListLength)

		// 集合操作
		api.GET("/set/:key", redisHandler.GetSet)
		api.GET("/set/:key/length", redisHandler.GetSetLength)

		// 哈希操作
		api.GET("/hash/:key", redisHandler.GetHash)
		api.GET("/hash/:key/length", redisHandler.GetHashLength)

		// 有序集合操作
		api.GET("/zset/:key", redisHandler.GetZSet)
		api.GET("/zset/:key/length", redisHandler.GetZSetLength)
	}

	// 静态文件服务
	r.NoRoute(gin.WrapH(handler.GetStaticHandler()))

	// 启动服务器
	go func() {
		time.Sleep(time.Second)
		openBrowser("http://localhost:8080")
	}()

	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
