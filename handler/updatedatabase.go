package handler

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/mayadata-io/ci-e2e-status/config"
)

// UpdateDatabase will update the latest pipelines detail and status
// TODO
func UpdateDatabase(gitlab config.Config) {
	// Read token environment variable
	token, ok := os.LookupEnv(token)
	if !ok {
		glog.Fatalf("TOKEN environment variable required")
	}
	update(gitlab, token)
	tick := time.Tick(10 * time.Minute)
	for range tick {
		update(gitlab, token)
	}
}

func update(gitLab config.Config, token string) {
	for _, project := range gitLab.Projects {
		for _, branch := range project.Branches {
			branchFormat := strings.Replace(branch.Name, "-", "_", -1)
			pipelineTable := fmt.Sprintf("%s_%s", project.Name, branchFormat)
			pipelineJobTable := fmt.Sprintf("%s_jobs", pipelineTable)
			getPlatformData(token, project.ID, branch.Name, pipelineTable, pipelineJobTable, branch.ReleaseTagJob)
		}
	}
}
