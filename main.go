package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//FileBak ...
type FileBak struct {
	Name     string
	FullName string
	Size     int64
	Date     time.Time
	BaseName string
	Period   string // формат База_ГодМесяц
	Deleted  bool
}

//FilesDel Список файлов для удаления
type FilesDel struct {
	files []FileBak
}

//GetSize Размер файлов для удаления
func (r *FilesDel) GetSize() (size int64) {
	size = 0
	for _, file := range r.files {
		if file.Deleted {
			size += file.Size
		}
	}
	return
}

//GetFile Количество файлов для удаления
func (r *FilesDel) GetFile() (kol int64) {
	kol = 0
	for _, file := range r.files {
		if file.Deleted {
			kol++
		}
	}
	return
}

const (
	configFileName = "files/config.yaml"
	logFileName    = "files/dir_file_dir.log"
)

var (
	fileListDel   FilesDel
	cfg           *Config
	dateEnd20Day  time.Time
	dateEnd3Month time.Time
	//Если старше 20 дней. Чётные числа удаляем + диапазон[11-20]
	filterDayDel = []int{2, 4, 6, 8, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 22, 24, 26, 28, 30}
	filePeriod   map[string]string
	messageMail  string
)

func main() {
	//загружаем конфиг
	var err error
	cfg, err = readConfig(configFileName)
	if err != nil {
		log.Fatalf("Не удалось загрузить %s: %s", configFileName, err)
	}
	//инициализируем логгеры
	if err := initLogger(cfg); err != nil {
		log.Fatalln(err)
	}

	fileListDel = FilesDel{}

	now := time.Now()
	dateEnd3Month = now.AddDate(0, -3, 0)  // - 3 месяцев
	dateEnd20Day = now.AddDate(0, -0, -20) // -20 дней
	filePeriod = make(map[string]string)
	//log.Trace("\nToday:", now, " dateEnd20Day:", dateEnd20Day," dateEnd3Month", dateEnd3Month)
	scanLoop(cfg)

	if cfg.SendEmail {
		log.Printf("Отправляем на адрес: %s сообщение: %s\n", cfg.ToEmail, messageMail)
		SendEmail("del_file_dir", cfg.ToEmail, "Результаты чистки диска на сервер", messageMail, "")
	}
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
		fileList := readDir(fullSrcDir)

		/*
			Формируем список файлов для удаления
			- файлы за последние 20 дней оставляем
			- чётные числа удаляем + диапазон[11-20]
			- оставляем последний файл в каждом периоде
		*/
		for _, file := range fileList {
			if len(cfg.MasksPattern) > 0 {
				// Проверяем по masks_pattern
				if matched, _ := cfg.matchParent(file.Name); !matched {
					//log.Trace("Пропускаем - ", file.Name)
					continue
				}
			} else {
				//Проверяем маски
				if matched, _ := cfg.match(file.Name); !matched {
					//log.Trace("Пропускаем - ", file.Name)
					continue
				}
			}

			// Маска исключения
			if matched, _ := cfg.matchExclude(file.Name); matched {
				//log.Trace("Пропускаем - ", file.Name)
				continue
			}

			// Получаем максимально возможную информацию из имени файла
			getInfoFromFile(&file)

			if dateEnd20Day.Before(file.Date) {
				//log.Trace("не удаляем(<20 дней) - ", file.Name, file.Date, dateEnd20Day)
				continue
			}

			if i := sort.SearchInts(filterDayDel, file.Date.Day()); i < len(filterDayDel) && filterDayDel[i] == file.Date.Day() {
				//log.Trace("Пропускаем - ", file.Name, file.Date.Day(), filterDayDel)
				continue
			}

			//log.Trace("OK - ", file.Name)
			file.Deleted = true
			fileListDel.files = append(fileListDel.files, file)
		}
		//Сортируем по имени базы и дате
		sort.Slice(fileListDel.files, func(i, j int) bool {
			switch strings.Compare(fileListDel.files[i].BaseName, fileListDel.files[j].BaseName) {
			case -1:
				return true
			case 1:
				return false
			}
			return fileListDel.files[i].Date.After(fileListDel.files[j].Date)
		})

		for index := 0; index < len(fileListDel.files); index++ {
			if dateEnd3Month.After(fileListDel.files[index].Date) {
				// Если нет файла в filePeriod то добавляем иначе удаляем из среза
				if _, ok := filePeriod[fileListDel.files[index].Period]; !ok {
					filePeriod[fileListDel.files[index].Period] = fileListDel.files[index].Name
					fileListDel.files[index].Deleted = false
					log.Trace(" ok - ", fileListDel.files[index].Name)
				}
			} else {
				log.Trace("до 3 мес. - ", fileListDel.files[index].Name)
			}
		}
		s := ""
		log.Info(fullSrcDir)
		messageMail += fullSrcDir + "\n"
		disk := DiskUsage(fullSrcDir)
		s = fmt.Sprintf("Свободно:     %8.2f GB", float64(disk.Free)/float64(GB))
		log.Info(s)
		messageMail += s + "\n"
		log.Printf("Файлов для удаления: %d. Размер: %8.2f GB", fileListDel.GetFile(), float64(fileListDel.GetSize())/float64(GB))
		messageMail += fmt.Sprintf("Файлов для удаления: %d. Размер: %8.2f GB\n", fileListDel.GetFile(), float64(fileListDel.GetSize())/float64(GB))
		for index := 0; index < len(fileListDel.files); index++ {
			if fileListDel.files[index].Deleted {
				s = fileListDel.files[index].Name + "- Удаляем"
				log.Info(s)
				messageMail += s + "\n"
				if err := removeFile(fileListDel.files[index].FullName); err != nil {
					log.Fatal(err)
				}
			}
		}
		disk = DiskUsage(fullSrcDir)
		s = fmt.Sprintf("Всего:        %8.2f GB", float64(disk.All)/float64(GB))
		log.Info(s)
		messageMail += s + "\n"
		s = fmt.Sprintf("Использовано: %8.2f GB", float64(disk.Used)/float64(GB))
		log.Info(s)
		messageMail += s + "\n"
		s = fmt.Sprintf("Свободно:     %8.2f GB", float64(disk.Free)/float64(GB))
		log.Info(s)
		messageMail += s + "\n"
	}
}

//readDir Читаем файлы из каталога и заносим их в fileList
func readDir(searchDir string) []FileBak {
	fileList := []FileBak{}
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
	return fileList
}

// удаление файла с диска
func removeFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// Получаем информацию из имени файла
func getInfoFromFile(file *FileBak) {
	//pattern := `(\d{4})_(\d{2})_(\d{2})`
	pattern := `(\w+)_backup_(\d\d\d\d)_(\d\d)_(\d\d)_\w+.bak`
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(file.Name)
	if match != nil {
		//log.Println(match[1], match[2], match[3], match[4])
		file.BaseName = match[1]
		year, _ := strconv.ParseInt(match[2], 10, 0)
		month, _ := strconv.ParseInt(match[3], 10, 0)
		day, _ := strconv.ParseInt(match[4], 10, 0)
		file.Date = time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
		file.Period = match[1] + "_" + match[2] + match[3]
		//fmt.Println(file.BaseName, file.Date, file.Period)
	}
}
