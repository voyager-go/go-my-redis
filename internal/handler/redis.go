package handler

import (
	"context"
	"fmt"
	"go-my-redis/internal/model"
	"go-my-redis/internal/service"
	"net/http"

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

	// 分割命令，保持引号内的内容完整
	parts := make([]string, 0)
	current := ""
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(request.Command); i++ {
		char := request.Command[i]

		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current += string(char)
			}
		} else if char == ' ' && !inQuotes {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

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

// GetList 获取列表数据
func (h *RedisHandler) GetList(c *gin.Context) {
	key := c.Param("key")
	list, err := h.redisService.GetList(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetListLength 获取列表长度
func (h *RedisHandler) GetListLength(c *gin.Context) {
	key := c.Param("key")
	length, err := h.redisService.GetListLength(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, length)
}

// GetSet 获取集合数据
func (h *RedisHandler) GetSet(c *gin.Context) {
	key := c.Param("key")
	set, err := h.redisService.GetSet(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, set)
}

// GetSetLength 获取集合长度
func (h *RedisHandler) GetSetLength(c *gin.Context) {
	key := c.Param("key")
	length, err := h.redisService.GetSetLength(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, length)
}

// GetHash 获取哈希数据
func (h *RedisHandler) GetHash(c *gin.Context) {
	key := c.Param("key")
	hash, err := h.redisService.GetHash(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hash)
}

// GetHashLength 获取哈希字段数
func (h *RedisHandler) GetHashLength(c *gin.Context) {
	key := c.Param("key")
	length, err := h.redisService.GetHashLength(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, length)
}

// GetZSet 获取有序集合数据
func (h *RedisHandler) GetZSet(c *gin.Context) {
	key := c.Param("key")
	zset, err := h.redisService.GetZSet(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, zset)
}

// GetZSetLength 获取有序集合长度
func (h *RedisHandler) GetZSetLength(c *gin.Context) {
	key := c.Param("key")
	length, err := h.redisService.GetZSetLength(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, length)
}
