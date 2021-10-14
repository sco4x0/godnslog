package main

import (
	"flag"
	Config "godnslog/Config"
	_ "godnslog/Database"
	Service "godnslog/Service"
)

func main() {
	config_file := flag.String("c", "./config.yml", "Config file path")
	flag.Parse()
	Config.SetConfigFile(*config_file)
	Config.Init()
	go Service.DnsServe()
	Service.HttpServe()
}
