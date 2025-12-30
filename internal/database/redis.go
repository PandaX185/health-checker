package database

import (
	"os"

	"github.com/go-redis/redis/v8"
)

var RdbInstance = redis.NewClient(
	&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	},
)
