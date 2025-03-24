package service

import (
	"context"
	"fmt"
	"go-my-redis/internal/model"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService() *RedisService {
	return &RedisService{}
}

func (s *RedisService) Connect(config model.RedisConfig) error {
	s.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()
	_, err := s.client.Ping(ctx).Result()
	return err
}

func (s *RedisService) GetKeys(pattern string) ([]string, error) {
	ctx := context.Background()
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *RedisService) GetKey(key string) (*model.RedisKey, error) {
	ctx := context.Background()

	// 获取key类型
	keyType, err := s.client.Type(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 获取TTL
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 根据类型获取值
	var value interface{}
	switch keyType {
	case "string":
		value, err = s.client.Get(ctx, key).Result()
	case "list":
		value, err = s.client.LRange(ctx, key, 0, -1).Result()
	case "set":
		value, err = s.client.SMembers(ctx, key).Result()
	case "hash":
		value, err = s.client.HGetAll(ctx, key).Result()
	case "zset":
		value, err = s.client.ZRange(ctx, key, 0, -1).Result()
	}

	if err != nil {
		return nil, err
	}

	return &model.RedisKey{
		Key:   key,
		Type:  keyType,
		Value: value,
		TTL:   ttl.Milliseconds(),
	}, nil
}

func (s *RedisService) SetKey(key string, value interface{}, ttl int64) error {
	ctx := context.Background()

	if ttl > 0 {
		return s.client.Set(ctx, key, value, time.Duration(ttl)*time.Millisecond).Err()
	}
	return s.client.Set(ctx, key, value, 0).Err()
}

func (s *RedisService) DeleteKey(key string) error {
	ctx := context.Background()
	return s.client.Del(ctx, key).Err()
}

// ExecuteCommand 执行 Redis 命令
func (s *RedisService) ExecuteCommand(ctx context.Context, parts ...string) (interface{}, error) {
	// 将 []string 转换为 []interface{}
	args := make([]interface{}, len(parts))
	for i, v := range parts {
		args[i] = v
	}
	return s.client.Do(ctx, args...).Result()
}

func (s *RedisService) GetType(key string) (string, error) {
	ctx := context.Background()
	return s.client.Type(ctx, key).Result()
}

func (s *RedisService) GetTTL(key string) (int64, error) {
	ctx := context.Background()
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int64(ttl.Seconds()), nil
}

func (s *RedisService) Expire(key string, seconds int64) error {
	ctx := context.Background()
	return s.client.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
}
