package storage

import (
	"fmt"
	"log"

	"github.com/crt379/svc-collector-grpc-gw/internal/config"

	"github.com/go-redis/redis"
)

var (
	WriteRedis *redis.Client
	ReadRedis  *redis.Client
)

func init() {
	var err error

	WriteRedis, err = NewRedisClient(
		config.AppConfig.Redis.Write.Host,
		config.AppConfig.Redis.Write.Port,
		config.AppConfig.Redis.Write.Password,
		config.AppConfig.Redis.Write.DB,
	)
	if err != nil {
		log.Panicf(err.Error())
	}

	ReadRedis, err = NewRedisClient(
		config.AppConfig.Redis.Read.Host,
		config.AppConfig.Redis.Read.Port,
		config.AppConfig.Redis.Read.Password,
		config.AppConfig.Log.Redis.DB,
	)
	if err != nil {
		log.Panicf(err.Error())
	}
}

func NewRedisClient(host, port, password string, Db int) (*redis.Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})
	if _, err := c.Ping().Result(); err != nil {
		return nil, err
	}

	return c, nil
}
