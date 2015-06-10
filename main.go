package main

import (
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"github.com/op/go-logging"

)

var log = logging.MustGetLogger("example")

type ServiceResponse struct {
	Message   string
	Timestamp int32
}

type ServiceHandler struct {
	Routes map[string]func(http.ResponseWriter, *http.Request)
}

func ( this *ServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := this.Routes[r.URL.String()]; ok {
		h(w, r)
		return
	}
	log.Error("Invalid path")
	fmt.Fprintf(w, "Invalid path")
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


func main() {

	var handler *ServiceHandler

	handler = &ServiceHandler{
		Routes: make(map[string]func(http.ResponseWriter, *http.Request)),
	}

	handler.Routes["/status"] = status


	server := http.Server{
		Addr:    ":8000",
		Handler: handler,
	}

	log.Info("Starting on :8000")
	server.ListenAndServe()

}
