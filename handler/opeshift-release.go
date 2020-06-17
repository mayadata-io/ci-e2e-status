package handler

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

// OpenshiftHandlerReleasee return eks pipeline data to /build path
func OpenshiftHandlerReleasee(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Openshiftdashboard{}
	err := QueryData(&datas, "release_pipeline_data", "release_jobs_data")
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
		return
	}
	out, err := json.Marshal(datas)
	if err != nil {
		http.Error(w, err.Error(), 500)
		glog.Error(err)
		return
	}
	w.Write(out)
}
