package toolbox

//import "github.com/callMe-Root/unbound-control-api/model"

import (

	"net/http"
	"os/exec"
	//"encoding/json"
	"log"
	"github.com/gorilla/mux"
)


func triggerUpdate(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	url := vars["url"]
	//target := vars["target"]

	//http.Get(url)
	curl := exec.Command("curl", url, "| bash")
	err := curl.Run()
	if  err == nil {
		w.WriteHeader(200)
	}else{
	log.Fatal(err)
	}

	
}