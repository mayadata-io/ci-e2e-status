package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"
	_ "github.com/lib/pq"
	"github.com/mayadata-io/ci-e2e-status/database"
	"github.com/mayadata-io/ci-e2e-status/handler"
	"github.com/mayadata-io/ci-e2e-status/router"

)

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	// Initailze Db connection
	database.InitDb()
	// Return value to /api path
	http.HandleFunc("/api/master/openshift", handler.OpenshiftHandlerMaster)
	http.HandleFunc("/api/master/build", handler.BuildhandlerMaster)
	http.HandleFunc("/api/release/openshift", router.OpenshiftHandlerRelease)
	glog.Infof("Listening on http://0.0.0.0:3000")

	// Trigger db update function
	go handler.UpdateDatabase()
	glog.Info(http.ListenAndServe(":"+"3000", nil))
}
