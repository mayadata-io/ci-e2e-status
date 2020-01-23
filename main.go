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
	http.HandleFunc("/packet/ultimate", handler.PacketHandlerUltimate)
	http.HandleFunc("/packet/penultimate", handler.PacketHandlerPenultimate)
	http.HandleFunc("/packet/antepenultimate", handler.PacketHandlerAntepenultimate)
	http.HandleFunc("/konvoy", handler.KonvoyHandler)
	http.HandleFunc("/openshift/release", handler.OpenshiftHandlerReleasee)

	glog.Infof("Listening on http://0.0.0.0:3000")

	// Trigger db update function
	go handler.UpdateDatabase()
	glog.Info(http.ListenAndServe(":"+"3000", nil))
}
