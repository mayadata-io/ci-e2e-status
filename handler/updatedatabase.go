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
	triggerType := []string{"master", "release"}
	// projects := []string{"maya", "jiva", "istgt", "zfs"}
	for _, t := range triggerType {
		// Fetch the e2e-openshift commit based pipeline
		if t == "master" {
			// go openshiftCommit(token, "e2e-openshift", "OpenEBS-base", "build_pipeline", "build_jobs")
		} else {
			releaseBranch(token, "e2e-openshift", "release-branch", "release_pipeline_data", "release_jobs_data")
		}
	}
	// Update the database, This wil run only first time
	// for _, project := range projects {
	// 	BuildData(token, project)
	// }
	// OpenshiftData(token, "openshift_pid", "openshift_pipeline", "openshift_jobs")
	// loop will iterate at every 2nd minute and update the database
	tick := time.Tick(10 * time.Minute)
	for range tick {
		// Fetch the e2e-openshift commit based pipeline
		for _, t := range triggerType {
			// Fetch the e2e-openshift commit based pipeline
			if t == "master" {
				// go openshiftCommit(token, "e2e-openshift", "OpenEBS-base", "build_pipeline", "build_jobs")
			} else {
				releaseBranch(token, "e2e-openshift", "release-branch", "release_pipeline_data", "release_jobs_data")
			}
		}
		// Fetch the pipeline detail of specified projects
		// for _, project := range projects {
		// 	BuildData(token, project)
		// }
		// // Fetch openshift pipeline
		// OpenshiftData(token, "openshift_pid", "openshift_pipeline", "openshift_jobs")
	}
}
