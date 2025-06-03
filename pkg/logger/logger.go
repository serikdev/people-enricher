package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger = *logrus.Entry

func NewLogger() Logger {
	log := logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg: "message",
		},
	})
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	log.SetOutput(os.Stdout)

	logg := logrus.NewEntry(log)
	return logg
}
