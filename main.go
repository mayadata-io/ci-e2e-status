package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"
	_ "github.com/lib/pq"
	"github.com/openebs/ci-e2e-dashboard-go-backend/database"
	"github.com/openebs/ci-e2e-dashboard-go-backend/handler"
)

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	// Initailze Db connection
	database.InitDb()
	// Return value to all / api path
	http.HandleFunc("/openshift", handler.OpenshiftHandler)
	http.HandleFunc("/build", handler.Buildhandler)
	glog.Infof("Listening on http://0.0.0.0:3000")

	// Trigger db update function
	go handler.UpdateDatabase()
	glog.Info(http.ListenAndServe(":"+"3000", nil))
}
