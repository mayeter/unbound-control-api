package internal

import (

	"net/http"


	"github.com/gorilla/mux"
)


func test(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["dir"]

	jsonResponse(w, 200, path)

}
