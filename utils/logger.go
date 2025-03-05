package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

func InitLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
	return logger
}
