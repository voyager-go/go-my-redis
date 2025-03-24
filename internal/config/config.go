package config

type Config struct {
	Port int
}

func NewConfig(port int) *Config {
	return &Config{
		Port: port,
	}
}
