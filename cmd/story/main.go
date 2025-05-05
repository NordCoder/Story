package main

import (
	"github.com/NordCoder/Story/config"
	"github.com/NordCoder/Story/internal/app"
	"github.com/NordCoder/Story/internal/logger"
	"go.uber.org/zap"
)

func main() {
	loggerConfig := config.NewLoggerConfig()

	log, err := logger.Init(loggerConfig)
	if err != nil {
		panic(err)
	}

	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil {
			log.Error("failed to sync logs", zap.Error(err))
		}
	}(log)

	httpCfg := config.NewHTTPConfig()

	app.Run(httpCfg, log)
}
