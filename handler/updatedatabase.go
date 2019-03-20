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
	OpenshiftData(token, "openshift_pid", "openshift_pipeline", "openshift_jobs")
	// loop will iterate at every 2nd minute and update the database
	tick := time.Tick(10 * time.Minute)
	for range tick {
		BuildData(token)
		OpenshiftData(token, "openshift_pid", "openshift_pipeline", "openshift_jobs")
	}
}
