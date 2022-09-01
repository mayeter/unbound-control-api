package toolbox

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"

	// "github.com/gorilla/mux"
	"github.com/callMe-Root/unbound-control-api/model"
)

func healthcheck(w http.ResponseWriter, r *http.Request) {

	//vars := mux.Vars(r)
	//cmd := vars["cmd"]
	//target := vars["target"]

	// services := ["bird", "unbound", "unbound-control-api"]
	hc := model.Healthcheck{}
	birdState := exec.Command("systemctl", "status", "bird")
	err := birdState.Run()
	if err == nil {
		hc.BirdState = "ok"
	} else {
		log.Fatal(err)
	}

	unboundState := exec.Command("systemctl", "status", "unbound")
	err = unboundState.Run()
	if err != nil {
		log.Fatal(err)
	}
	hc.UnboundState = "ok"

	apiState := exec.Command("systemctl", "status", "unbound-control-api")
	err = apiState.Run()
	if err != nil {
		log.Fatal(err)
	}
	hc.ApiState = "ok"

	json.NewEncoder(w).Encode(hc)

}
