package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var ServerConfig *Config

type Config struct {
	Name      string `json:"ServerName"`
	Host      string `json:"Host"`
	IPVersion string `json:"IPVersion"`
	Port      int    `json:"Port"`
	MaxConn   int    `json:"MaxConn"`
}

func InitConfig(filename string) error {
	var (
		content []byte
		err     error
	)
	if content, err = ioutil.ReadFile(filename); err != nil {
		return err
	}
	// 解析配置
	if err = json.Unmarshal(content, &ServerConfig); err != nil {
		return err
	}
	return err
}

func init() {
	fmt.Println("config")
}
