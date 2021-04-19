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
			case "nativek8s":
				for _, nativeBranch := range database.NativeBranch {
					go getPlatformData(token, NATIVEK8SID, strings.Replace(nativeBranch, "_", "-", -1), fmt.Sprintf("%s_%s", platform, nativeBranch), fmt.Sprintf("%s_%s_%s", platform, nativeBranch, "jobs"))
				}
			}

		}
	}
	// go getPlatformData(token, NATIVEK8SID, "release-branch", pipelineTable, pipelineTableJobs)

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
					case "nativek8s":
						for _, nativeBranch := range database.NativeBranch {
							go getPlatformData(token, NATIVEK8SID, b, fmt.Sprintf("%s_%s", platform, nativeBranch), fmt.Sprintf("%s_%s_%s", platform, nativeBranch, "jobs"))
						}
					}
				}

			}
		}

	}
}
