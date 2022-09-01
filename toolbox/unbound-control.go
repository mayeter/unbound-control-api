package toolbox

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/callMe-Root/unbound-control-api/model"
	"github.com/gorilla/mux"
)

//"io/ioutil"

func callUC(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	cmd := vars["cmd"]
	//target := vars["target"]

	res, err := exec.Command("sudo", "unbound-control", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}

	dataArray := strings.Split(string(res), "\n")
	if len(dataArray) > 0 {
		dataArray = dataArray[:len(dataArray)-1]
	}

	var jsonArray []model.ForwardZone

	for _, element := range dataArray {
		var iterator = model.ForwardZone{}

		iterator.Zone = strings.Fields(element)[0]
		iterator.Class = strings.Fields(element)[1]
		iterator.Type = strings.Fields(element)[2]

		for i := 3; i < len(strings.Fields(element)); i++ {
			iterator.Forwarders = append(iterator.Forwarders, strings.Fields(element)[i])
		}

		jsonArray = append(jsonArray, iterator)

	}

	json.NewEncoder(w).Encode(jsonArray)

}
