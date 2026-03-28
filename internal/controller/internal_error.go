package controller

import (
	"log"
	"net/http"

	"recipebox-backend-go/internal/utils"
)

func writeInternalError(w http.ResponseWriter, r *http.Request, publicMessage string, err error) {
	log.Printf("%s %s failed: %v", r.Method, r.URL.Path, err)
	utils.Error(w, http.StatusInternalServerError, publicMessage)
}
