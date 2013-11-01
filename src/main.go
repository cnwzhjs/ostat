/**
 * Created with IntelliJ IDEA.
 * User: Tony
 * Date: 13-10-31
 * Time: 下午2:16
 * To change this template use File | Settings | File Templates.
 */
package main

import (
	"log"
	"net/http"
	"fmt"
	"encoding/json"
	"time"
	"github.com/stretchr/goweb"
	"github.com/stretchr/goweb/context"
)

type ServerStatus struct {
	Address string
	StartTime time.Time
	Services []string
	Depends []string
}

func main() {
	serverStatus := ServerStatus{":8080", time.Now(), []string {"stat"}, []string {}}

	log.Printf("Starting server at %s...", serverStatus.Address)

	server := &http.Server {
		Addr: serverStatus.Address,
		Handler: goweb.DefaultHttpHandler(),
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	goweb.Map("GET", "/", func (c context.Context) {
			JsonResponse(c.HttpResponseWriter(), serverStatus, nil)
		})

	log.Fatal(server.ListenAndServe())
}

type Response struct {
	Result interface{}
	Error error
}

func JsonResponse(w http.ResponseWriter, data interface{}, error error) {
	response := Response{data, error}
	result, error := json.Marshal(response)

	if error != nil {
		JsonResponse(w, nil, error)
	} else {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(result))
	}
}
