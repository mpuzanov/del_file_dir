package main

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-yaml/yaml"
)

// Config - структура для считывания конфигурационного файла
type Config struct {
	LogLevel     string           `yaml:"loglevel"`
	SrcDirs      []string         `yaml:"src_dirs"`
	ExcludeMasks []string         `yaml:"exclude_masks,omitempty"`
	Masks        []string         `yaml:"masks,omitempty"`
	MasksPattern []string         `yaml:"masks_pattern,omitempty"`
	ToEmail      string           `yaml:"toEmail,omitempty"`
	SendEmail    bool             `yaml:"sendEmail,omitempty"`
	SettingsSMTP EmailCredentials `yaml:"settingsSMTP,omitempty"`
}

func readConfig(ConfigName string) (x *Config, err error) {
	var file []byte
	if file, err = ioutil.ReadFile(ConfigName); err != nil {
		return nil, err
	}
	x = new(Config)
	if err = yaml.Unmarshal(file, x); err != nil {
		return nil, err
	}
	if x.LogLevel == "" {
		x.LogLevel = "Trace"
	}
	return x, nil
}

//проверяет подходит ли файл под маски данного правила
//возвращает список масок
func (r *Config) match(srcFile string) (bool, []string) {
	var masks []string
	for _, mask := range r.Masks {
		//fmt.Println(mask, srcFile)
		matched, err := filepath.Match(strings.ToLower(mask), strings.ToLower(srcFile))
		if err != nil {
			log.Printf("Ошибка проверки MASK (%s). %s", mask, err)
			continue
		}
		if matched {
			masks = append(masks, mask)
		}
	}
	if len(masks) == 0 {
		return false, masks
	}
	return true, masks
}

// Проверяем маски исключения
// возвращает список масок
func (r *Config) matchExclude(srcFile string) (bool, []string) {
	var masks []string
	for _, mask := range r.ExcludeMasks {
		//fmt.Println(mask, srcFile)
		matched, err := filepath.Match(strings.ToLower(mask), strings.ToLower(srcFile))
		if err != nil {
			log.Printf("Ошибка проверки MASK (%s). %s", mask, err)
			continue
		}
		if matched {
			masks = append(masks, mask)
		}
	}
	if len(masks) == 0 {
		return false, masks
	}
	return true, masks
}

func (r *Config) matchParent(srcFile string) (bool, []string) {
	var masks []string
	for _, mask := range r.MasksPattern {
		re := regexp.MustCompile(mask)
		//fmt.Println(mask, srcFile, re.MatchString(srcFile))
		if re.MatchString(srcFile) {
			masks = append(masks, mask)
		}
	}
	if len(masks) == 0 {
		return false, masks
	}
	return true, masks
}
