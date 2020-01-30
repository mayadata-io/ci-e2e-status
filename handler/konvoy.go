package handler

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

// KonvoyHandler return packet pipeline data to /packet path
func KonvoyHandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Openshiftdashboard{}
	err := QueryData(&datas, "konvoy_pipeline", "konvoy_jobs")
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
	}
	out, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
	}
	w.Write(out)
}
