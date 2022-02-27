package config

import (
	"19PMI/19PMI/logs"
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"
)

var logger *logs.Logger

type ApplicationConfig struct {
	DBPath       string        `json:"db_path" default:"/opt/data/source.db"`
	ServerConfig *ServerConfig `json:"server" required:"true"`
}

type ServerConfig struct {
	Port int `json:"port" default:"8080"`
}

/**
 * Setup singleton
 */
var once sync.Once
var cInstance *ApplicationConfig

func InitApplicationConfiguration(configFilePath string) {
	once.Do(
		func() {
			loggerInstance := logs.GetLogger().SetSource("config")
			logger = &loggerInstance

			//defer logger.TimeTrack(time.Now(), "init")

			instance, err := getApplicationConfigurationInstance(configFilePath)
			if err != nil {
				panic(err)
			}

			cInstance = instance
		},
	)
}

func GetApplicationConfiguration() *ApplicationConfig {
	if cInstance == nil {
		err := errors.New("need init application configuration first")
		if logger != nil {
			logger.Err(err)
		}

		panic(err)
	}

	return cInstance
}

func getApplicationConfigurationInstance(configFilePath string) (instance *ApplicationConfig, err error) {
	var configurationBytes []byte
	configurationBytes, err = ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}

	if logger != nil {
		instance, err = ParseConfig(configurationBytes)
		if err != nil {
			return
		}

		var configBytes []byte
		configBytes, err = json.MarshalIndent(
			instance,
			"",
			"  ",
		)

		if err != nil {
			logger.Error().Msg(err.Error())

			return
		}

		logger.LogConfig(configBytes)
	}

	return
}
