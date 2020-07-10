package handler

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// DeletePipeline for delete a pipeline
func DeletePipeline(w http.ResponseWriter, r *http.Request) {
	allowedHeaders := "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token"
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
	w.Header().Set("Access-Control-Expose-Headers", "Authorization") // Double check it's a post request being made

	if r.Method == "DELETE" {
		r.ParseForm()
		platform := r.Form.Get("platform")
		PID := r.Form.Get("pid")
		fmt.Fprintf(w, " method : "+r.Method+" platform : "+platform+" ID : "+PID)
		switch platform {
		case "openshift":
			err := PipelineDeleteQuery("release_pipeline_data", PID)
			if err != nil {
				fmt.Fprintf(w, "unable to find the pipeline : %s", PID)
				return
			}
			fmt.Fprintf(w, "successfully Removed %s pipeline from %s platform .", PID, platform)
		case "nativek8s":
			err := PipelineDeleteQuery("nativek8s_pipeline", PID)
			if err != nil {
				fmt.Fprintf(w, "unable to find the pipeline : %s", PID)
				return
			}
			fmt.Fprintf(w, "successfully Removed %s pipeline from %s platform .", PID, platform)
		case "konvoy":
			err := PipelineDeleteQuery("konvoy_pipeline", PID)
			if err != nil {
				fmt.Fprintf(w, "unable to find the pipeline : %s", PID)
				return
			}
			fmt.Fprintf(w, "successfully Removed %s pipeline from %s platform .", PID, platform)
		case "packetAntepenultimate":
			err := PipelineDeleteQuery("packet_pipeline_k8s_antepenultimate", PID)
			if err != nil {
				fmt.Fprintf(w, "unable to find the pipeline : %s", PID)
				return
			}
			fmt.Fprintf(w, "successfully Removed %s pipeline from %s platform .", PID, platform)
		case "packetPenultimate":
			err := PipelineDeleteQuery("packet_pipeline_k8s_penultimate", PID)
			if err != nil {
				fmt.Fprintf(w, "unable to find the pipeline : %s", PID)
				return
			}
			fmt.Fprintf(w, "successfully Removed %s pipeline from %s platform .", PID, platform)
		case "packetUltimate":
			err := PipelineDeleteQuery("packet_pipeline_k8s_ultimate", PID)
			if err != nil {
				fmt.Fprintf(w, "unable to find the pipeline : %s", PID)
				return
			}
			fmt.Fprintf(w, "successfully Removed %s pipeline from %s platform .", PID, platform)
		default:
			fmt.Fprintf(w, "Pls check the Platform , %s Platform not found", platform)
		}

	} else {
		fmt.Fprintf(w, "Use Request method as Delete . Now using : "+r.Method)
	}

}

// PipelineDeleteQuery querying delete pipeline
func PipelineDeleteQuery(platform string, pid string) error {

	queryDelete := fmt.Sprintf("DELETE FROM %s WHERE id=%s", platform, pid)
	glog.Infoln("query : ", queryDelete)
	execQuery, err := database.Db.Query(queryDelete)
	if err != nil {
		return err
	}
	defer execQuery.Close()
	return nil
}
