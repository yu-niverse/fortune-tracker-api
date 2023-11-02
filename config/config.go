package config

import (
	"fmt"
	"github.com/spf13/viper"
)

var Viper *viper.Viper

// Load the config file from config/app.env to Viper
func LoadConfig() *viper.Viper {
	vp := viper.New()
	vp.SetConfigName("app")
	vp.SetConfigType("env")
	vp.AddConfigPath("config")
	vp.AutomaticEnv()
	
	if err := vp.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", vp.ConfigFileUsed())
		Viper = vp
		return vp
	} else {
		fmt.Println("Error loading config file:", err)
		return nil
	}
}