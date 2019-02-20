package controller

import (
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
)

type telemetryController struct {
}

func (t telemetryController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/telemetry/stream", t.handleTelemetry)
}

func (t telemetryController) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	methodName := "handleTelemetry"
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		go CustomLog(methodName+": Reading request failed.", ErrorSeverity)
	}

	go CustomLog(methodName+": Request: "+string(dump), DebugSeverity)
}
