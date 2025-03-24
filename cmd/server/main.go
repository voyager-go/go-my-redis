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
		api.POST("/disconnect", redisHandler.Disconnect)
		api.POST("/command", redisHandler.ExecuteCommand)

		// 键管理
		api.GET("/keys", redisHandler.GetKeys)
		api.GET("/key/:key", redisHandler.GetKey)
		api.POST("/key", redisHandler.SetKey)
		api.DELETE("/key/:key", redisHandler.DeleteKey)
		api.GET("/type/:key", redisHandler.GetType)
		api.GET("/ttl/:key", redisHandler.GetTTL)
		api.POST("/expire", redisHandler.Expire)

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

	log.Fatal(r.Run(":8080"))
}
