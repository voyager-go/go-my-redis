package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	rdb *redis.Client
}

func (s *RedisService) ExecuteCommand(command string) (interface{}, error) {
	// 解析命令和参数
	args := strings.Fields(command)
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := strings.ToLower(args[0])
	params := args[1:]

	// 处理带引号的参数
	processedParams := make([]string, 0)
	currentParam := ""
	inQuotes := false
	quoteChar := rune(0)

	for _, param := range params {
		if !inQuotes {
			if strings.HasPrefix(param, "\"") || strings.HasPrefix(param, "'") {
				inQuotes = true
				quoteChar = rune(param[0])
				currentParam = param[1:]
			} else {
				processedParams = append(processedParams, param)
			}
		} else {
			if strings.HasSuffix(param, string(quoteChar)) {
				currentParam += " " + param[:len(param)-1]
				processedParams = append(processedParams, currentParam)
				inQuotes = false
				currentParam = ""
			} else {
				currentParam += " " + param
			}
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("unclosed quotes")
	}

	// 根据命令类型处理
	switch cmd {
	case "lpush":
		if len(processedParams) < 2 {
			return nil, fmt.Errorf("lpush requires at least 2 arguments")
		}
		key := processedParams[0]
		values := make([]interface{}, len(processedParams[1:]))
		for i, v := range processedParams[1:] {
			values[i] = v
		}
		return s.rdb.LPush(context.Background(), key, values...).Result()
	case "rpush":
		if len(processedParams) < 2 {
			return nil, fmt.Errorf("rpush requires at least 2 arguments")
		}
		key := processedParams[0]
		values := make([]interface{}, len(processedParams[1:]))
		for i, v := range processedParams[1:] {
			values[i] = v
		}
		return s.rdb.RPush(context.Background(), key, values...).Result()
	case "lpop":
		if len(processedParams) != 1 {
			return nil, fmt.Errorf("lpop requires exactly 1 argument")
		}
		return s.rdb.LPop(context.Background(), processedParams[0]).Result()
	case "rpop":
		if len(processedParams) != 1 {
			return nil, fmt.Errorf("rpop requires exactly 1 argument")
		}
		return s.rdb.RPop(context.Background(), processedParams[0]).Result()
	case "lrange":
		if len(processedParams) != 3 {
			return nil, fmt.Errorf("lrange requires exactly 3 arguments")
		}
		key := processedParams[0]
		start, err := strconv.ParseInt(processedParams[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid start index")
		}
		stop, err := strconv.ParseInt(processedParams[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid stop index")
		}
		return s.rdb.LRange(context.Background(), key, start, stop).Result()
	default:
		return nil, fmt.Errorf("unsupported command: %s", cmd)
	}
}
