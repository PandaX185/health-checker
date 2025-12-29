package config

import (
	"sync"

	"go.uber.org/zap"
)

var logger *zap.Logger
var loggerOnce = &sync.Once{}

func NewLogger(env string) *zap.Logger {
	loggerOnce.Do(
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
