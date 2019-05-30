package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"
	_ "github.com/lib/pq"
	"github.com/mayadata-io/ci-e2e-status/database"
	"github.com/mayadata-io/ci-e2e-status/handler"
)

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	// Initailze Db connection
	database.InitDb()
	// Return value to all / api path
	http.HandleFunc("/packet/v11", handler.PacketHandlerV11)
	http.HandleFunc("/packet/v12", handler.PacketHandlerV12)
	http.HandleFunc("/packet/v13", handler.PacketHandlerV13)
	http.HandleFunc("/build", handler.Buildhandler)
	glog.Infof("Listening on http://0.0.0.0:3000")

	// Trigger db update function
	go handler.UpdateDatabase()
	glog.Info(http.ListenAndServe(":"+"3000", nil))
}
