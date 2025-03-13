package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	server := InitWebServer()

	initViper()
	server.Run(":8080")
}

func initViper() {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
