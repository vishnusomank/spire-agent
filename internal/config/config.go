package config

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	ConfigPath string
}

func Default() *Config {
	return &Config{}
}

func (cfg *Config) LoadConfigFile() {
	viper.SetConfigFile(cfg.ConfigPath)
	fmt.Printf("viper.GetViper().ConfigFileUsed(): %v\n", viper.GetViper().ConfigFileUsed())
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal().Msgf("Reading config file[%s] failed. %s", viper.ConfigFileUsed(), err)
		return
	}
}

func (cfg *Config) PrintConfig() {
	allKeys := viper.AllKeys()
	configArr := []string{}

	for _, key := range allKeys {
		value := viper.Get(key)
		configArr = append(configArr, fmt.Sprintf("%s=%v", key, value))
	}

	log.Info().Msgf("Application configuration - [%s]", strings.Join(configArr, ", "))
}
