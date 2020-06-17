package handler

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

// PacketHandlerAntepenultimate return packet pipeline data to /packet path
func PacketHandlerAntepenultimate(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Openshiftdashboard{}
	err := QueryData(&datas, "packet_pipeline_k8s_antepenultimate", "packet_jobs_k8s_antepenultimate")
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

// PacketHandlerPenultimate return packet pipeline data to /packet path
// TODO
func PacketHandlerPenultimate(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Openshiftdashboard{}
	err := QueryData(&datas, "packet_pipeline_k8s_penultimate", "packet_jobs_k8s_penultimate")
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

// PacketHandlerUltimate return packet pipeline data to /packet path
// TODO
func PacketHandlerUltimate(w http.ResponseWriter, r *http.Request) {
	// Allow cross origin request
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	datas := Openshiftdashboard{}
	err := QueryData(&datas, "packet_pipeline_k8s_ultimate", "packet_jobs_k8s_ultimate")
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
