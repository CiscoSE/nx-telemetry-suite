package controller

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// BasePath is use to search file in gopath directory
var BasePath = os.Getenv("GOPATH") + "/src/github.com/CiscoSE/nx-telemetry-suite"

// ErrorSeverity a constant use for log purposes
const ErrorSeverity = "ERROR"

// DebugSeverity a constant use for log purposes
const DebugSeverity = "DEBUG"

var (
	indexCtl     indexController
	homeCtl      homeController
	telemetryCtl telemetryController
)

// Startup associates controllers with templates and routes
func Startup(templates map[string]*template.Template, r *mux.Router) {

	// Handle web server mappings

	// Home & Index
	indexCtl.template = templates["index.html"]
	indexCtl.registerRoutes(r)

	homeCtl.template = templates["home.html"]
	homeCtl.registerRoutes(r)

	// Telemetry streams
	telemetryCtl.registerRoutes(r)

	// Public assets and configs
	r.PathPrefix("/assets/").Handler(http.FileServer(http.Dir(BasePath + "/public")))

}

// CreateDirIfNotExist creates directories if not present
func CreateDirIfNotExist(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				go CustomLog("CreateDirIfNotExist (create directory): "+err.Error(), ErrorSeverity)
			}
		} else {
			go CustomLog("CreateDirIfNotExist (check directory): "+err.Error(), ErrorSeverity)
		}
	}
}
