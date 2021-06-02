package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/mayadata-io/ci-e2e-status/config"
	"github.com/mayadata-io/ci-e2e-status/database"
	"github.com/mayadata-io/ci-e2e-status/handler"
)

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	// Read confugurations from /config/config.json
	gitlab := config.ReadConfig()
	// Initailze Db connection
	database.InitDb(gitlab)

	r := mux.NewRouter()
	r.HandleFunc("/status", handler.StatusGitLab)
	r.HandleFunc("/config", config.ViewConfig)
	r.HandleFunc("/{platform}/{branch}", handler.OpenshiftHandlerReleasee)
	r.HandleFunc("/{platform}/{branch}/pipeline/{id}", handler.GetPipelineDataAPI)
	r.HandleFunc("/{platform}/{branch}/job/{id}/raw", handler.GetJobLogs)
	r.HandleFunc("/recent", handler.HandleRecentPipelines)

	// Trigger db update function
	go handler.UpdateDatabase(gitlab)
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:3000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	glog.Infof("Listening on http://0.0.0.0:3000")
	log.Fatal(srv.ListenAndServe())
}
