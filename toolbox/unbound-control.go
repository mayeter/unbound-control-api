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

func lookup(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	recordName, ok := params["record_name"]
	if ok {
		if len(recordName) == 1 {
			res, err := exec.Command("sudo", "unbound-control", "lookup", recordName[0]).Output()
			if err != nil {
				var warningMessage model.WarningMessage
				warningMessage.Message = err.Error()
				json.NewEncoder(w).Encode(warningMessage)
				log.Fatal(err)
			}
			dataArray := strings.Split(string(res), "\n")
			var jsonArray model.Lookup
			for _, element := range dataArray {
				jsonArray.LookupServer = append(jsonArray.LookupServer, element)
			}
			json.NewEncoder(w).Encode(jsonArray)

		} else {
			var warningMessage model.WarningMessage
			warningMessage.Message = "You should use record_name parameter only once!!!"
			json.NewEncoder(w).Encode(warningMessage)
		}

		var warningMessage model.WarningMessage
		warningMessage.Message = "query result is not ok"
		json.NewEncoder(w).Encode(warningMessage)

	}
}
