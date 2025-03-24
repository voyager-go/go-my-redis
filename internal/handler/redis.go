package handler

import (
	"context"
	"fmt"
	"go-my-redis/internal/model"
	"go-my-redis/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type RedisHandler struct {
	redisService *service.RedisService
}

func NewRedisHandler() *RedisHandler {
	return &RedisHandler{
		redisService: service.NewRedisService(),
	}
}

func (h *RedisHandler) Connect(c *gin.Context) {
	var config model.RedisConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.redisService.Connect(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connected successfully"})
}

func (h *RedisHandler) GetKeys(c *gin.Context) {
	pattern := c.DefaultQuery("pattern", "*")
	fmt.Println("pattern", pattern)
	keys, err := h.redisService.GetKeys(pattern)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rdb_keys": keys})
}

func (h *RedisHandler) GetKey(c *gin.Context) {
	key := c.Param("key")
	redisKey, err := h.redisService.GetKey(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, redisKey)
}

func (h *RedisHandler) SetKey(c *gin.Context) {
	var redisKey model.RedisKey
	if err := c.ShouldBindJSON(&redisKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.redisService.SetKey(redisKey.Key, redisKey.Value, redisKey.TTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key set successfully"})
}

func (h *RedisHandler) DeleteKey(c *gin.Context) {
	key := c.Param("key")
	if err := h.redisService.DeleteKey(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key deleted successfully"})
}

func (h *RedisHandler) ExecuteCommand(c *gin.Context) {
	var request struct {
		Command string `json:"command"`
	}

	// 解析请求体
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 分割命令
	parts := strings.Fields(request.Command)
	if len(parts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No command provided"})
		return
	}

	// 执行命令
	ctx := context.Background()
	result, err := h.redisService.ExecuteCommand(ctx, parts...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回结果
	response := map[string]interface{}{
		"result": result,
	}
	c.JSON(http.StatusOK, response)
}

func (h *RedisHandler) GetType(c *gin.Context) {
	key := c.Param("key")
	keyType, err := h.redisService.GetType(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keyType)
}

func (h *RedisHandler) GetTTL(c *gin.Context) {
	key := c.Param("key")
	ttl, err := h.redisService.GetTTL(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ttl)
}

func (h *RedisHandler) Expire(c *gin.Context) {
	var request struct {
		Key     string `json:"key"`
		Seconds int64  `json:"seconds"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.redisService.Expire(request.Key, request.Seconds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "TTL set successfully"})
}

func (h *RedisHandler) Disconnect(c *gin.Context) {
	if err := h.redisService.Disconnect(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Disconnected successfully"})
}
