package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type telemetryController struct {
}

func (t telemetryController) registerRoutes(r *mux.Router) {
	r.HandleFunc("/", t.handleTelemetry)
}

func (t telemetryController) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	methodName := "handleTelemetry"
	// If get, redirect to home page. If post, save to influxDB
	switch r.Method {
	case http.MethodGet:
		http.Redirect(w, r, "/web/", http.StatusSeeOther)
	case http.MethodPost:
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			go CustomLog(methodName+": Reading request failed.", ErrorSeverity)
		}
		go CustomLog(methodName+": Telemetry Payload: "+string(b[:]), DebugSeverity)
		go t.addJSONToInfluxDB(b)
	}

}

// Add a json payload (parsed to []byte) to influxdb
func (t telemetryController) addJSONToInfluxDB(data []byte) {
	methodName := "addJSONToInfluxDB"

	var iMetrics interface{}

	// Unmarshall json
	err := json.Unmarshal(data, &iMetrics)
	if err != nil {
		go CustomLog(methodName+": cannot unmarshall json. "+err.Error(), ErrorSeverity)
		return
	}

	influxURL := fmt.Sprintf("http://%s:%s/write?db=%s", os.Getenv("INFLUX_HOST"), os.Getenv("INFLUX_PORT"), os.Getenv("INFLUX_DB"))
	// Assert metrics
	metrics := iMetrics.(map[string]interface{})

	// Add tags. Assuming that all telemtry sent will have these values
	tags := "node_id_str=" + metrics["node_id_str"].(string) + ",encoding_path=" + metrics["encoding_path"].(string)
	// Get time
	epochTimeStr := metrics["collection_end_time"].(string)
	// TODO: Support for subsecond epochs
	epochTime, err := strconv.ParseInt(epochTimeStr, 10, 64)
	if err != nil {
		go CustomLog(methodName+": epoch time conversion failed. "+err.Error(), ErrorSeverity)
		return
	}

	// load data to database
	go loadMapToInfluxDB(influxURL, metrics["data"].(map[string]interface{}), tags, epochTime)

}

// Recursive function to traverse an unknown json, adding all keys and values
func loadMapToInfluxDB(influxURL string, dataMap map[string]interface{}, tags string, epochTime int64) {
	methodName := "loadMapToInfluxDB"

	for key := range dataMap {
		tmpType := reflect.ValueOf(dataMap[key])
		if tmpType.Kind() == reflect.Map {
			// If value is a map, call function again
			loadMapToInfluxDB(influxURL, dataMap[key].(map[string]interface{}), tags, epochTime)
		} else {
			payload := strings.NewReader("")
			// If it is not a map, add to influxDB. Assuming is either string or int.
			switch tmpType.Kind() {
			case reflect.Int:
				payload = strings.NewReader(key + "," + tags + " value=" + strconv.Itoa(dataMap[key].(int)) + " " + strconv.FormatInt(epochTime, 10))
			case reflect.Float32:
			case reflect.Float64:
				payload = strings.NewReader(key + "," + tags + " value=" + strconv.FormatFloat(dataMap[key].(float64), 'f', -1, 64) + " " + strconv.FormatInt(epochTime, 10))
			default:
				// Defaults to string
				payload = strings.NewReader(key + "," + tags + " value=\"" + dataMap[key].(string) + "\" " + strconv.FormatInt(epochTime, 10))
			}

			req, err := http.NewRequest("POST", influxURL, payload)

			if err != nil {
				go CustomLog(methodName+": influxDB write failed. "+err.Error(), ErrorSeverity)
			}

			res, err := http.DefaultClient.Do(req)

			if err != nil {
				go CustomLog(methodName+": influxDB write failed. "+err.Error(), ErrorSeverity)
			}

			if res.StatusCode > 299 || res.StatusCode < 200 {
				body, _ := ioutil.ReadAll(res.Body)
				go CustomLog(methodName+": Request to influxDB failed: Status code="+strconv.Itoa(res.StatusCode)+" - Message="+string(body), ErrorSeverity)
			}
		}

	}
}
