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

type Healthcheck struct {
	UnboundState string `json:"unbound"`
	BirdState    string `json:"bird"`
	ApiState     string `json:"api"`
}
