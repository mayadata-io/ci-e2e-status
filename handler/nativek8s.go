package handler

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

// Nativek8sHandler return packet pipeline data to /packet path
func Nativek8sHandler(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Openshiftdashboard{}
	err := QueryData(&datas, "nativek8s_pipeline", "nativek8s_jobs")
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
