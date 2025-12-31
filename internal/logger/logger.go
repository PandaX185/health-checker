package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	log  *zap.Logger
	once sync.Once
)

func New(env string) *zap.Logger {
	once.Do(func() {
		var err error
		if env == "production" {
			log, err = zap.NewProduction()
		} else {
			log, err = zap.NewDevelopment()
		}

		if err != nil {
			panic(err)
		}
	})

	return log
}

func Get() *zap.Logger {
	return log
}
