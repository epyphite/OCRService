package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/epyphite/OCRService/pkg/models"
	"github.com/epyphite/OCRService/pkg/service/ocr"
	webapi "github.com/epyphite/OCRService/pkg/service/web"
	"github.com/epyphite/OCRService/pkg/utils"
)

var rootCmd = &cobra.Command{
	Use:   "OCRService",
	Short: "OCR Service",
	Long:  ``,
	RunE:  ocrService,
}

//Execute will run the desire module command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var url string
var configFile string
var webserver bool

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Specify a configuration file")
	rootCmd.PersistentFlags().BoolVar(&webserver, "webserver", false, "Start a rest webserver")
}

func ocrService(cmd *cobra.Command, args []string) error {

	var config models.Config
	var srv ocr.Service
	var err error

	if configFile == "" {
		config, err = utils.LoadConfigurationDefaults()

	} else {
		config, err = utils.LoadConfiguration(configFile)

	}
	srv.Config = config

	fmt.Println(config)

	if configFile != "" {
		log.Println("Initialize ...")
		srv.Init()
		log.Println("Process")

		if webserver {
			webagent, err := webapi.NewWebAgent(config)
			if err != nil {
				log.Fatalln("Error on newebagent call ", err)
			}
			log.Println("Starting Web Server in", config.WebAddress, config.WebPort)

			go utils.HandleSignal()
			webagent.StartServer()
			return err
		}
		if config.ReadQueue == "yes" {
			go srv.ReadResponse()
		}
		if config.ProcessInput == "yes" {
			srv.Process()
		}

	}
	return err
}
