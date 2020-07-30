package webapi

import (
	"log"
	"net/http"

	"github.com/epyphite/OCRService/pkg/models"
	"github.com/gorilla/handlers"

	constants "github.com/epyphite/OCRService/pkg/constants"
	webapi "github.com/epyphite/OCRService/pkg/service/web/app"
)

//APIOne main structure
type APIOne struct {
	webconfig models.Config
}

//NewWebAgent // creates a mew instace \of web one
func NewWebAgent(config models.Config) (APIOne, error) {
	var APIOne APIOne
	APIOne.webconfig = config
	return APIOne, nil
}

//StartServer Starts the server using the variable sip and port, creates anew instance.
func (W *APIOne) StartServer() {
	log.Println("Version : " + constants.BuildVersion)
	log.Println("Build Time : " + constants.BuildTime)
	handler := W.New()

	http.ListenAndServe(W.webconfig.WebAddress+":"+W.webconfig.WebPort, handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"*"}),
	)(handler))
}

//New creates a new handler
func (W *APIOne) New() http.Handler {

	app, err := webapi.NewApp(W.webconfig)

	if err != nil {
		log.Fatalln("Error creating API ")
		return nil
	}

	api := app.Mux.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/liveness", app.Liveness)

	return &app
}
