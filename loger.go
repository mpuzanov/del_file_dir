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

	checkMaxFileSize(cfg, logFileName)

	if cfg.LogToFile {
		file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			log.Println("logToFile: ", logFileName)
			log.Out = file
		} else {
			log.Info("Failed to log to file, using default stderr")
		}
	}
	log.Println("log.Level:", log.Level)
	return nil
}

func checkMaxFileSize(cfg *Config, logFileName string) error {
	file, err := os.Stat(logFileName)
	if err != nil {
		return err
	}
	maxFileSize := cfg.MaxFileSize * 1024 * 1024
	if file.Size() > maxFileSize {
		log.Println("Очищаем файл ", logFileName)
		f, err := os.Create(logFileName)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}
