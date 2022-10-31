package model

// import (

// 	"os/exec"
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"io/ioutil"
// 	"strings"

// 	"github.com/gorilla/mux"
// )
type WarningMessage struct {
	Message string `json:"message"`
}

type ForwardZone struct {
	Zone       string   `json:"zone"`
	Class      string   `json:"-"`
	Type       string   `json:"-"`
	Forwarders []string `json:"forwarders"`
}

type CheckCheck struct {
	HealthCheck struct {
		Unbound    string `json:"unbound"`
		ControlAPI string `json:"control-api"`
	} `json:"healthcheck"`
	VersionCheck struct {
		Unbound    string `json:"unbound"`
		Config     string `json:"config"`
		ControlAPI string `json:"control-api"`
	} `json:"versioncheck"`
}

type Lookup struct {
	LookupServer []string `json:"server"`
}
