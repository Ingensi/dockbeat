package redis

import "time"

type redisConfig struct {
	Host              string        `config:"host"`
	Port              int           `config:"port"`
	Password          string        `config:"password"`
	Db                int           `config:"db"`
	DbTopology        int           `config:"db_topology"`
	Timeout           time.Duration `config:"timeout"`
	Index             string        `config:"index"`
	ReconnectInterval int           `config:"reconnect_interval"`
	DataType          string        `config:"datatype"`
}

var (
	defaultConfig = redisConfig{
		DbTopology:        1,
		Timeout:           5 * time.Second,
		ReconnectInterval: 1,
	}
)
