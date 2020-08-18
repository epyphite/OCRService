package MainWebApi

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/gorilla/sessions"

	models "github.com/epyphite/OCRService/pkg/models"
	"github.com/epyphite/OCRService/pkg/service/ocr"
	c1 "github.com/epyphite/OCRService/pkg/service/web/constants"
)

//JResponse create a trscture to respond json
type JResponse struct {
	ResponseCode string
	Message      string
	ResponseData []byte
}

//MainWebAPI PHASE
type MainWebAPI struct {
	Mux    *mux.Router
	Log    *log.Logger
	Config models.Config
}

//NewApp create a new object for the App.
func NewApp(config models.Config) (MainWebAPI, error) {
	var err error
	var wapp MainWebAPI

	mux := mux.NewRouter().StrictSlash(true)

	log := log.New(os.Stdout, "API", log.LstdFlags)
	wapp.Mux = mux
	wapp.Config = config
	wapp.Log = log

	if err != nil {
		log.Println(err)
	}
	return wapp, err
}

func (a *MainWebAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Mux.ServeHTTP(w, r)
}

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := c1.Store.Get(r, "ocrService-session")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil
	}
	return session
}

//Liveness just keeps the connection alive
func (a *MainWebAPI) Liveness(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var response JResponse

	response.Message = "Process Alive"
	response.ResponseCode = "200"
	response.ResponseData = nil
	js, err := json.Marshal(response)
	if err != nil {
		log.Println()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "Application/json")
	w.Write(js)
}

//ProcessInput will process TODO
func (a *MainWebAPI) ProcessInput(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var config models.Config
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&config)
	log.Println("Request Body ", err)
	log.Println(r.Body)
	var srv ocr.Service

	srv.Config = config
	finished := make(chan bool)

	srv.Init()
	log.Println("Processing Configuration file ")
	log.Println(config)
	if config.ProcessInput == "yes" {
		log.Println("Launching read folder")
		go srv.Process(finished)

	}
	if config.ReadQueue == "yes" {
		log.Println("Launching Queue")
		go srv.ReadResponse(finished)
	}

	var response JResponse
	response.Message = "Process Started "
	response.ResponseCode = "201"
	response.ResponseData = nil
	js, err := json.Marshal(response)
	if err != nil {
		log.Println()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "Application/json")
	w.Write(js)
}
