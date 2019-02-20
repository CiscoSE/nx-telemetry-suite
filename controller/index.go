package controller

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

type indexController struct {
	template *template.Template
}

func (i indexController) registerRoutes(r *mux.Router) {
	r.NotFoundHandler = http.HandlerFunc(i.redirectHome)
	r.HandleFunc("/web/", i.handleIndex)
	r.HandleFunc("/web/index", i.handleIndex)
}

func (i indexController) handleIndex(w http.ResponseWriter, r *http.Request) {

	i.template.Execute(w, nil)
}

func (i indexController) redirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/web/", http.StatusSeeOther)
}
