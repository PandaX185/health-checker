package database

import "github.com/go-redis/redis/v8"

var Rdb = redis.NewClient(
	&redis.Options{
		Addr: "localhost:6379",
	},
)
