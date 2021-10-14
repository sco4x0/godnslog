package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type ConfigProps struct {
	Listen     string `yaml:"listen"`
	Domain     string `yaml:"domain"`
	HttpPort   int    `yaml:"http_port"`
	HttpView   bool   `yaml:"http_view"`
	HttpOffset int    `yaml:"http_offset"`
	Token      string `yaml:"token"`
}

var Config ConfigProps
var filepath string

func Init() {
	log.Printf("Load config file ...")
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic("Read config file error !")
	}
	if err = yaml.Unmarshal(content, &Config); err != nil {
		panic("Config file parse error !")
	}
}

func SetConfigFile(file string) {
	filepath = file
}
