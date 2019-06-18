package router

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/handler"
)

// OpenshiftHandlerRelease return eks pipeline data to /build path
func OpenshiftHandlerRelease(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := handler.Builddashboard{}
	err := handler.QueryOpenshiftReleaseData(&datas, "release_pipeline_data", "release_jobs_data")
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
