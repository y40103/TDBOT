package utils

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"runtime"
)

func GetRedis(dbname string) *redis.Client {
	url := fmt.Sprintf("redis://:@localhost:6379/%v", dbname)
	opt, err := redis.ParseURL(url)
	opt.PoolSize = runtime.GOMAXPROCS(runtime.NumCPU() * 5)
	opt.MinIdleConns = runtime.GOMAXPROCS(runtime.NumCPU() * 3)
	if err != nil {
		panic(err)
	}

	return redis.NewClient(opt)
}
