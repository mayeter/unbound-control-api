package toolbox

import (
	//"encoding/json"
	"fmt"
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
	a.Router.HandleFunc("/api/unbound-control/{cmd}", callUC).Methods("GET")
	http.Handle("/", a.Router)
}

func (a *App) Run(addr string) {
	fmt.Println("listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))

}
