package main

import (
	"go-my-redis/internal/handler"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	redisHandler := handler.NewRedisHandler()

	// 配置CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API路由组
	api := r.Group("/api")
	{
		// Redis连接管理
		api.POST("/connect", redisHandler.Connect)
		api.GET("/keys", redisHandler.GetKeys)
		api.GET("/key/:key", redisHandler.GetKey)
		api.POST("/key", redisHandler.SetKey)
		api.DELETE("/key/:key", redisHandler.DeleteKey)
		api.GET("/type/:key", redisHandler.GetType)
		api.GET("/ttl/:key", redisHandler.GetTTL)
		api.POST("/command", redisHandler.ExecuteCommand)
		api.POST("/expire", redisHandler.Expire)
	}

	log.Fatal(r.Run(":8080"))
}
