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

	checkMaxFileSize(cfg)

	if cfg.LogToFile {
		file, err := os.OpenFile(cfg.LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			log.Println("logToFile => ", cfg.LogFileName) // выводим на консоль имя файла для логов
			log.Out = file
		} else {
			log.Infof("Failed to log to file <%s>, using default stderr", cfg.LogFileName)
		}
	}
	log.Println("=============================================================")
	log.Println("log.Level:", log.Level)
	return nil
}

func checkMaxFileSize(cfg *Config) error {
	file, err := os.Stat(cfg.LogFileName)
	if err != nil {
		return err
	}
	maxFileSize := cfg.MaxFileSize * 1024 * 1024
	if file.Size() > maxFileSize {
		log.Println("Очищаем файл ", cfg.LogFileName)
		f, err := os.Create(cfg.LogFileName)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}
