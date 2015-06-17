package main

import (
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"github.com/op/go-logging"
	"dishash/storage"
	"regexp"
//"strings"
)

var log = logging.MustGetLogger("example")

var stor *storage.Storage = storage.Init(5000000000)

type ServiceResponse struct {
	Message   string
	Timestamp int32
}

type ServiceRequestRoute struct {
	Path   string
	Method string
}

type ServiceHandler struct {
	Routes map[ServiceRequestRoute]func(http.ResponseWriter, *http.Request)
}

func ( this *ServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	filter := regexp.MustCompile("^/([a-zA-Z0-9]+)/([a-zA-Z0-9]+)$")
	url_parsed := filter.FindStringSubmatch(r.URL.Path)

	key := ServiceRequestRoute{
		Path: "/"+url_parsed[1],
		Method: r.Method,
	}

	if h, ok := this.Routes[key]; ok {
		h(w, r)
		return
	}

	log.Error("Invalid path" +r.URL.Path)
	fmt.Fprintf(w, "Invalid path "+r.URL.Path)
}


func status(w http.ResponseWriter, r *http.Request) {

	response := ServiceResponse{
		Message:  "Ok",
		Timestamp: int32(time.Now().Unix()),
	}

	json_result, err := json.Marshal(response)

	if err==nil {
		fmt.Fprintf(w, string(json_result))
		log.Info("Request processed /status")
	} else {
		fmt.Fprintf(w, "JSON marshaling error!")
		log.Error("JSON marshaling error!")
	}

}

func setKey(w http.ResponseWriter, r *http.Request) {

	response := ServiceResponse{
		Message:  "Error",
		Timestamp: int32(time.Now().Unix()),
	}

	key := r.URL.Path[len("/keys/"):]
	value := r.FormValue("value")

	if value != "" {
		stor.Set(key, value)
		response = ServiceResponse{
			Message:  "Ok",
			Timestamp: int32(time.Now().Unix())    }
	}


	json_result, _ := json.Marshal(response)
	fmt.Fprintf(w, string(json_result))
	log.Info("Request processed /keys/:key")


}

func getKey(w http.ResponseWriter, r *http.Request) {

	key := r.URL.Path[len("/keys/"):]
	value := stor.Get(key)
	response := ServiceResponse{
		Message:  value,
		Timestamp: int32(time.Now().Unix()),
	}

	json_result, _ := json.Marshal(response)
	fmt.Fprintf(w, string(json_result))

}


func main() {

	var handler *ServiceHandler

	handler = &ServiceHandler{
		Routes: make(map[ServiceRequestRoute]func(http.ResponseWriter, *http.Request)),
	}

	handler.Routes[ServiceRequestRoute{Path: "/status", Method: "GET"}] = status
	handler.Routes[ServiceRequestRoute{Path: "/keys", Method: "GET"}] = getKey
	handler.Routes[ServiceRequestRoute{Path: "/keys", Method: "POST"}] = setKey

	server := http.Server{
		Addr:    ":8000",
		Handler: handler,
	}

	log.Info("Starting on :8000")
	server.ListenAndServe()

}
