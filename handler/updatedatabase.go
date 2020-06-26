package handler

import (
	"os"
	"time"

	"github.com/golang/glog"
)

// UpdateDatabase will update the latest pipelines detail and status
// TODO
func UpdateDatabase() {
	// Read token environment variable
	token, ok := os.LookupEnv(token)
	if !ok {
		glog.Fatalf("TOKEN environment variable required")
	}
	// Update the database, This wil run only first time
	k8sVersion := []string{"ultimate", "penultimate", "antepenultimate"}
	for _, k8sVersion := range k8sVersion {
		branch := "k8s-" + k8sVersion
		pipelineTable := "packet_pipeline_k8s_" + k8sVersion
		jobTable := "packet_jobs_k8s_" + k8sVersion
		getPlatformData(token, PACKETID, branch, pipelineTable, jobTable) //e2e-packet
	}
	go getPlatformData(token, KONVOYID, "release-branch", "konvoy_pipeline", "konvoy_jobs")                // e2e-konvoy
	go getPlatformData(token, OPENSHIFTID, "release-branch", "release_pipeline_data", "release_jobs_data") //e2e-openshift
	go getPlatformData(token, NATIVEK8SID, "release-branch", "nativek8s_pipeline", "nativek8s_jobs")       //e2e-openshift

	// loop will iterate at every 2nd minute and update the database
	tick := time.Tick(2 * time.Minute)
	for range tick {
		k8sVersion := []string{"ultimate", "penultimate", "antepenultimate"}
		for _, k8sVersion := range k8sVersion {
			branch := "k8s-" + k8sVersion
			pipelineTable := "packet_pipeline_k8s_" + k8sVersion
			jobTable := "packet_jobs_k8s_" + k8sVersion
			getPlatformData(token, PACKETID, branch, pipelineTable, jobTable) //e2e-packet
		}
		go getPlatformData(token, KONVOYID, "release-branch", "konvoy_pipeline", "konvoy_jobs")                // e2e-konvoy
		go getPlatformData(token, OPENSHIFTID, "release-branch", "release_pipeline_data", "release_jobs_data") //e2e-openshift
		go getPlatformData(token, NATIVEK8SID, "release-branch", "nativek8s_pipeline", "nativek8s_jobs")       //e2e-openshift
	}

}
