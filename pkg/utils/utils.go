package utils

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/epyphite/OCRService/pkg/constants"
	"github.com/epyphite/OCRService/pkg/models"
)

//LoadConfiguration returns the read Configuration and error while reading.
func LoadConfiguration(file string) (models.Config, error) {
	var config models.Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return config, err
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config, err
}

//LoadConfigurationDefaults will load default values from constants
func LoadConfigurationDefaults() (models.Config, error) {
	var config models.Config
	var err error

	config.CloudSourceStorage = constants.SourceFolder
	config.SourceType = constants.SourceType
	config.Debug = "no"
	config.EnableCloud = "yes"

	return config, err
}

//HandleSignal a master function to handle interrupt
func HandleSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitChan := make(chan int)
	go func() {
		for {
			s := <-signalChan
			switch s {
			// kill -SIGHUP XXXX
			case syscall.SIGHUP:
				log.Println("hungup")

			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				log.Println("Warikomi")
				exitChan <- 0

			// kill -SIGTERM XXXX
			case syscall.SIGTERM:
				log.Println("force stop")
				exitChan <- 0

			// kill -SIGQUIT XXXX
			case syscall.SIGQUIT:
				log.Println("stop and core dump")
				exitChan <- 0

			default:
				log.Println("Unknown signal.")
				exitChan <- 1
			}
		}
	}()

	code := <-exitChan
	os.Exit(code)
}
