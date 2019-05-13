package monitor

import (
	redis "gopkg.in/redis.v2"
)

func getNewRedisConnection() *redis.Client {
	opt := &redis.Options{Addr: "127.0.0.1:6379", DB: 0, Network: "tcp"}
	c := redis.NewClient(opt)
	c.Auth("passme")
	return c
}
