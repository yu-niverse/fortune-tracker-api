package config

import (
	"github.com/spf13/viper"
)

var Viper *viper.Viper

// Load the config file from config/app.env to Viper
func LoadConfig() *viper.Viper {
	vp := viper.New()
	vp.AutomaticEnv()
	Viper = vp
	return vp
}