package model

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type RedisKey struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
	TTL   int64       `json:"ttl"`
}

// ZSetMember 表示有序集合的成员
type ZSetMember struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}
