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
	BuildData(token)
	k8sVersion := []string{"v15", "v14", "v13"}
	for _, k8sVersion := range k8sVersion {
		columnName := "packet_" + k8sVersion + "_pid"
		pipelineTable := "packet_pipeline_" + k8sVersion
		jobTable := "packet_jobs_" + k8sVersion
		go PacketData(token, columnName, pipelineTable, jobTable)
	}
	go KonvoyData(token, "konvoy_pid", "konvoy_pipeline", "konvoy_jobs")
	go releaseBranch(token, "e2e-openshift", "release-branch", "release_pipeline_data", "release_jobs_data")

	// loop will iterate at every 2nd minute and update the database
	tick := time.Tick(2 * time.Minute)
	for range tick {
		BuildData(token)
		for _, k8sVersion := range k8sVersion {
			columnName := "packet_" + k8sVersion + "_pid"
			pipelineTable := "packet_pipeline_" + k8sVersion
			jobTable := "packet_jobs_" + k8sVersion
			go PacketData(token, columnName, pipelineTable, jobTable)
		}
		go KonvoyData(token, "konvoy_pid", "konvoy_pipeline", "konvoy_jobs")
		go releaseBranch(token, "e2e-openshift", "release-branch", "release_pipeline_data", "release_jobs_data")

	}
}
