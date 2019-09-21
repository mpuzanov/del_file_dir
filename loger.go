package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func initLogger(cfg *Config) error {
	log.Formatter = new(logrus.TextFormatter)
	log.Formatter.(*logrus.TextFormatter).TimestampFormat = "02-01-2006 15:04:05"
	log.Formatter.(*logrus.TextFormatter).FullTimestamp = true

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Error(err)
	}
	log.SetLevel(level)

	if cfg.LogToFile {
		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			log.Infoln("logToFile: ", logFileName)
			log.Out = file
		} else {
			log.Info("Failed to log to file, using default stderr")
		}
	}
	log.Infoln("log.Level:", log.Level)
	return nil
}
