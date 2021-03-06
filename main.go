package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/CiscoSE/nx-telemetry-suite/controller"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	templates := populateTemplates()

	controller.Startup(templates, r)
	log.Println("Listening in http://0.0.0.0:" + os.Getenv("APP_WEB_PORT") + "/web/")
	err := http.ListenAndServe(":"+os.Getenv("APP_WEB_PORT"), r)
	if err != nil {
		controller.CustomLog("Failed to start web server: "+err.Error(), controller.ErrorSeverity)
		os.Exit(1)
	}
}

func populateTemplates() map[string]*template.Template {
	result := make(map[string]*template.Template)
	templateBasePath := controller.BasePath + "/htmlTemplates"
	layout := template.Must(template.ParseFiles(templateBasePath + "/_layout.html"))
	template.Must(
		layout.ParseFiles(templateBasePath + "/_default_menu.html"))
	dir, err := os.Open(templateBasePath + "/content")
	if err != nil {
		panic("Failed to open template blocks directory: " + err.Error())
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		panic("Failed to read contents of content directory: " + err.Error())
	}
	for _, fi := range fis {
		// DEBUG:
		//log.Print("Reading html template file " + fi.Name())
		f, err := os.Open(templateBasePath + "/content/" + fi.Name())
		if err != nil {
			panic("Failed to open template '" + fi.Name() + "'")
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			panic("Failed to read content from file '" + fi.Name() + "'")
		}
		f.Close()
		tmpl := template.Must(layout.Clone())
		_, err = tmpl.Parse(string(content))
		if err != nil {
			panic("Failed to parse contents of '" + fi.Name() + "' as template")
		}
		result[fi.Name()] = tmpl
	}
	return result
}
