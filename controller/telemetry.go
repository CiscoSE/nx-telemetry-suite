package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	influxDBClient "github.com/influxdata/influxdb1-client"
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

	// Create influxDB URL
	host, err := url.Parse(fmt.Sprintf("http://%s:%s", os.Getenv("INFLUX_HOST"), os.Getenv("INFLUX_PORT")))
	if err != nil {
		go CustomLog(methodName+": parse URL failed. "+err.Error(), ErrorSeverity)
		return
	}
	// Create client connection
	con, err := influxDBClient.NewClient(influxDBClient.Config{URL: *host})
	if err != nil {
		go CustomLog(methodName+": influxDB connection failed. "+err.Error(), ErrorSeverity)
		return
	}

	// Assert metrics
	metrics := iMetrics.(map[string]interface{})

	// Add tags. Assuming that all telemtry sent will have these values
	tags := map[string]string{
		"node_id_str":   metrics["node_id_str"].(string),
		"encoding_path": metrics["encoding_path"].(string),
	}

	// Get time
	epochTimeStr := metrics["collection_end_time"].(string)
	// TODO: Support for subsecond epochs
	epochTime, err := strconv.ParseInt(epochTimeStr[:len(epochTimeStr)-3], 10, 64)
	if err != nil {
		go CustomLog(methodName+": epoch time conversion failed. "+err.Error(), ErrorSeverity)
		return
	}

	// load data to database
	go loadMapToInfluxDB(con, metrics["data"].(map[string]interface{}), tags, epochTime)

}

// Recursive function to traverse an unknown json, adding all keys and values
func loadMapToInfluxDB(con *influxDBClient.Client, dataMap map[string]interface{}, tags map[string]string, epochTime int64) {
	methodName := "loadMapToInfluxDB"
	var fields map[string]interface{}

	for key := range dataMap {
		tmpType := reflect.ValueOf(dataMap[key])
		if tmpType.Kind() == reflect.Map {
			// If value is a map, call function again
			loadMapToInfluxDB(con, dataMap[key].(map[string]interface{}), tags, epochTime)
		} else {
			// If it is not a map, add to influxDB. Assuming is either string or int.
			if tmpType.Kind() == reflect.Int {
				fields = map[string]interface{}{
					"value": dataMap[key].(int64),
				}
			} else {
				// Defaults to string
				fields = map[string]interface{}{
					"value": dataMap[key],
				}
			}
			pts := make([]influxDBClient.Point, 1)
			pts[0] = influxDBClient.Point{
				Measurement: key,
				Tags:        tags,
				Fields:      fields,
				Time:        time.Unix(epochTime, 0),
				Precision:   "s",
			}

			bps := influxDBClient.BatchPoints{
				Points:   pts,
				Database: os.Getenv("INFLUX_DB"),
			}
			// Make the request
			_, err := con.Write(bps)
			if err != nil {
				go CustomLog(methodName+": influxDB write failed. "+err.Error(), ErrorSeverity)
			}
		}

	}
}
