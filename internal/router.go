package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

func (a *App) Init() {

	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/test/{dir}", test).Methods("GET", "OPTIONS")
	a.Router.PathPrefix("/").Handler(http.FileServer(http.Dir("../client/")))
}

func (a *App) Run(addr string) {
	fmt.Println("listening on " + addr)

	// Where ORIGIN_ALLOWED is like `scheme://dns[:port]`, or `*` (insecure)
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "OPTIONS"})

	log.Fatal(http.ListenAndServe(addr, handlers.CORS(originsOk, headersOk, methodsOk)(a.Router)))

}

func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
// func test(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	path := vars["dir"]

// 	jsonResponse(w, 200, path)

// }
