package main

import (
	"net/http"

	"github.com/golang/glog"
	_ "github.com/lib/pq"
	"github.com/openebs/ci-e2e-dashboard-go-backend/database"
	"github.com/openebs/ci-e2e-dashboard-go-backend/handler"
)

func main() {
	// Initailze Db connection
	database.InitDb()

	// Return value to all / api path
	http.HandleFunc("/gke", handler.Gkehandler)
	// http.HandleFunc("/aws", handler.Awshandler)
	// http.HandleFunc("/gcp", handler.Gcphandler)
	// http.HandleFunc("/aks", handler.Akshandler)
	// http.HandleFunc("/packet", handler.Packethandler)
	// http.HandleFunc("/eks", handler.Ekshandler)
	http.HandleFunc("/build", handler.Buildhandler)
	glog.Infof("Listening on http://localhost:3000")

	// Trigger db update function
	go handler.UpdateDatabase()
	glog.Info(http.ListenAndServe("localhost:3000", nil))
}
