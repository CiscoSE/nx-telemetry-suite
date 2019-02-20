package controller

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

type homeController struct {
	template *template.Template
}

func (h homeController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/ng/home", h.handleHome)
}

func (h homeController) handleHome(w http.ResponseWriter, r *http.Request) {
	h.template.Execute(w, nil)
}
