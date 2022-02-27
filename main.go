package main

import (
	"19PMI/19PMI/config"
	"19PMI/19PMI/logs"
	"19PMI/19PMI/service"
	"flag"
	"fmt"
	"math/rand"
	"time"
)

var logger *logs.Logger

// Called automatically before main func
func init() {
	configFilePath := ""
	logFileName := ""
	flag.StringVar(
		&configFilePath,
		"configFilePath",
		"./config/config.json",
		"Application configuration file path",
	)
	flag.StringVar(
		&logFileName,
		"logFileName",
		"",
		"File name for logging",
	)
	flag.Parse()

	logs.SetLogFileName(logFileName)
	config.InitApplicationConfiguration(configFilePath)

	loggerInstance := logs.GetLogger().SetSource("main")
	logger = &loggerInstance

	rand.Seed(time.Now().UnixNano())
}

func main() {
	go func() {
		webService := service.GetWebService()
		err := webService.Run()
		if err != nil {
			logger.Error().Msg(err.Error())
		}
	}()

	logger.Info().Msg("started")

	var input string
	_, _ = fmt.Scanln(&input)
}
