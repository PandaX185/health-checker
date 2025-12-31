package logger

import (
	"go.uber.org/zap"
)

func New(env string) *zap.Logger {
	var (
		log *zap.Logger
		err error
	)

	if env == "production" {
		log, err = zap.NewProduction()
	} else {
		log, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(err)
	}

	return log
}
