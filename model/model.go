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

type ForwardZone struct {
	Zone       string   `json:"zone"`
	Class      string   `json:"-"`
	Type       string   `json:"-"`
	Forwarders []string `json:"forwarders"`
}

type CheckCheck struct {
	HealthCheck struct {
		Unbound    string `json:"unbound"`
		Bird       string `json:"bird"`
		ControlAPI string `json:"control-api"`
	} `json:"healthcheck"`
	VersionCheck struct {
		Unbound    string `json:"unbound"`
		Config     string `json:"config"`
		Bird       string `json:"bird"`
		ControlAPI string `json:"control-api"`
	} `json:"versioncheck"`
}
