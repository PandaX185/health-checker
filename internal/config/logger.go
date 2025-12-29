package config

import (
	"sync"

	"go.uber.org/zap"
)

var logger *zap.Logger
var once sync.Once

func NewLogger(env string) *zap.Logger {
	once.Do(
		func() {
			var (
				err error
			)

			if env == "production" {
				logger, err = zap.NewProduction()
			} else {
				logger, err = zap.NewDevelopment()
			}

			if err != nil {
				panic(err)
			}

		},
	)
	return logger
}
