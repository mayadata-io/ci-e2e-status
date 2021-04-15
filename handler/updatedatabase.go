package handler

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/database"
)

// UpdateDatabase will update the latest pipelines detail and status
// TODO
func UpdateDatabase() {
	// Read token environment variable
	token, ok := os.LookupEnv(token)
	if !ok {
		glog.Fatalf("TOKEN environment variable required")
	}

	for _, platform := range database.Platform {
		for _, branch := range database.Branch {
			b := strings.Replace(branch, "_", "-", -1)
			pipelineTable := fmt.Sprintf("%s_%s", platform, branch)
			pipelineTableJobs := fmt.Sprintf("%s_jobs", pipelineTable)
			// glog.Infoln(fmt.Sprintf("\n branch: %s \n pipelineTableJobs : %s \n  Token : %s \n ", b, pipelineTableJobs, token))
			switch platform {
			case "openshift":
				go getPlatformData(token, OPENSHIFTID, b, pipelineTable, pipelineTableJobs)
			case "konvoy":
				go getPlatformData(token, KONVOYID, b, pipelineTable, pipelineTableJobs)
			}

		}
	}

	tick := time.Tick(2 * time.Minute)
	// // branch := []string{}
	for range tick {
		for _, platform := range database.Platform {
			for _, branch := range database.Branch {
				b := strings.Replace(branch, "_", "-", -1)
				pipelineTable := fmt.Sprintf("%s_%s", platform, branch)
				pipelineTableJobs := fmt.Sprintf("%s_jobs", pipelineTable)
				if gitLabStatus(BaseURL) {
					switch platform {
					case "openshift":
						go getPlatformData(token, OPENSHIFTID, b, pipelineTable, pipelineTableJobs)
					case "konvoy":
						go getPlatformData(token, KONVOYID, b, pipelineTable, pipelineTableJobs)
					}
				}

			}
		}

	}

	// Update the database, This wil run only first time
	// k8sVersion := []string{"ultimate", "penultimate", "antepenultimate"}
	// for _, k8sVersion := range k8sVersion {
	// 	branch := "k8s-" + k8sVersion
	// 	pipelineTable := "packet_pipeline_k8s_" + k8sVersion
	// 	jobTable := "packet_jobs_k8s_" + k8sVersion
	// 	getPlatformData(token, PACKETID, branch, pipelineTable, jobTable) //e2e-packet
	// }
	// go getPlatformData(token, KONVOYID, "release-branch", "konvoy_pipeline", "konvoy_jobs")                // e2e-konvoy
	// go getPlatformData(token, OPENSHIFTID, "release-branch", "release_pipeline_data", "release_jobs_data") //e2e-openshift
	// go getPlatformData(token, NATIVEK8SID, "release-branch", "nativek8s_pipeline", "nativek8s_jobs")       //e2e-nativek8s

	// // loop will iterate at every 2nd minute and update the database
	// tick := time.Tick(2 * time.Minute)
	// // branch := []string{}
	// for range tick {
	// 	k8sVersion := []string{"ultimate", "penultimate", "antepenultimate"}
	// 	for _, k8sVersion := range k8sVersion {
	// 		branch := "k8s-" + k8sVersion
	// 		pipelineTable := "packet_pipeline_k8s_" + k8sVersion
	// 		jobTable := "packet_jobs_k8s_" + k8sVersion
	// 		getPlatformData(token, PACKETID, branch, pipelineTable, jobTable) //e2e-packet
	// 	}
	// 	go getPlatformData(token, KONVOYID, "release-branch", "konvoy_pipeline", "konvoy_jobs")                // e2e-konvoy
	// 	go getPlatformData(token, OPENSHIFTID, "release-branch", "release_pipeline_data", "release_jobs_data") //e2e-openshift
	// 	go getPlatformData(token, NATIVEK8SID, "release-branch", "nativek8s_pipeline", "nativek8s_jobs")       //e2e-nativek8s
	// }

}
