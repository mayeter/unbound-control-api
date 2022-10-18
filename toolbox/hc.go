package toolbox

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"

	"github.com/callMe-Root/unbound-control-api/model"
)

func healthcheck(w http.ResponseWriter, r *http.Request) {

	hc := model.CheckCheck{}
	birdState := exec.Command("systemctl", "status", "bird")
	err := birdState.Run()
	if err == nil {
		hc.HealthCheck.Bird = "ok"
	} else {
		log.Fatal(err)
	}

	unboundState := exec.Command("systemctl", "status", "unbound")
	err = unboundState.Run()
	if err != nil {
		log.Fatal(err)
	}
	hc.HealthCheck.Unbound = "ok"

	apiState := exec.Command("systemctl", "status", "unbound-control-api")
	err = apiState.Run()
	if err != nil {
		log.Fatal(err)
	}
	hc.HealthCheck.ControlAPI = "ok"

	unboundVersion := exec.Command("unbound", "-V | head -n1")
	err = unboundVersion.Run()
	if err != nil {
		log.Fatal(err)
	}
	hc.VersionCheck.Unbound = "ok"

	configVersion := exec.Command("head", "-n1 /etc/unbound/unbound.conf")
	err = configVersion.Run()
	if err != nil {
		log.Fatal(err)
	}
	hc.VersionCheck.Config = "ok"

	hc.VersionCheck.ControlAPI = "v1.0.1"

	json.NewEncoder(w).Encode(hc)

}
