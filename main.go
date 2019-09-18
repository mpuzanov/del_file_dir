package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

//FileBak ...
type FileBak struct {
	Name     string
	FullName string
	Size     int64
	Date     time.Time
}

var (
	configFileName = "config.yaml"
	fileList       []FileBak
	cfg            *Config
)

func main() {
	//загружаем конфиг
	var err error
	cfg, err = readConfig(configFileName)
	if err != nil {
		log.Fatalf("Не удалось загрузить %s: %s", configFileName, err)
	}
	fileList = []FileBak{}
	scanLoop(cfg)
}

// scanLoop просматривает конфиг и для каждого каталога-источника
func scanLoop(cfg *Config) {
	for _, srcDir := range cfg.SrcDirs {
		fullSrcDir := srcDir
		abspath, err := filepath.Abs(fullSrcDir)
		if err != nil {
			log.Println("Ошибка вычисления абсолютного пути", srcDir, err)
			continue
		}
		fullSrcDir = abspath
		log.Println("Сканируем каталог", fullSrcDir)

		//читаем содержимое каталога
		readDir(fullSrcDir)

		for _, file := range fileList {
			//Проверяем маски
			if matched, _ := cfg.match(file.Name); !matched {
				fmt.Println("Пропускаем", file.Name)
				continue
			}
			fmt.Println(file)
		}
	}
}

func readDir(searchDir string) {

	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			file := FileBak{}
			file.Name = f.Name()
			file.Size = f.Size()
			file.FullName = path
			file.Date = f.ModTime()

			fileList = append(fileList, file)
		}
		return nil
	})
}

