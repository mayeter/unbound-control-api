package toolbox

import (
	"encoding/json"
	//"encoding/json"
	"fmt"
	//"io/ioutil"
	"log"
	"net/http"

	//"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

func (a *App) Init() {

	a.Router = mux.NewRouter()
	//a.Router.HandleFunc("/message", returnMessage).Methods("POST")
	a.Router.HandleFunc("/api/updater", triggerUpdate).Methods("POST")
	a.Router.HandleFunc("/api/healthcheck", healthcheck).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/callUC/{cmd}", callUC).Methods("GET")

	a.Router.HandleFunc("/api/unbound-control/lookup", lookup).Methods("GET").Queries("record_name")
	a.Router.HandleFunc("/api/unbound-control/start", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/stop", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/restart", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/dump_cache", maintenance).Methods("GET")

	a.Router.HandleFunc("/api/unbound-control/flush", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/flush_zone", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/flush_negative", maintenance).Methods("GET")

	a.Router.HandleFunc("/api/unbound-control/list_forwards", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/list_local_zones", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/list_local_data", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/view_list_local_zones", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/view_list_local_data", maintenance).Methods("GET")

	a.Router.HandleFunc("/api/unbound-control/forward_add", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/forward_remove", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/local_data", maintenance).Methods("GET")
	a.Router.HandleFunc("/api/unbound-control/local_data_remove", maintenance).Methods("GET")

	http.Handle("/", a.Router)
}

func (a *App) Run(addr string) {
	fmt.Println("listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))

}

func maintenance(w http.ResponseWriter, r *http.Request) {

	maintenanceNote := "this endpoint is under maintenance"

	json.NewEncoder(w).Encode(&maintenanceNote)

}
