package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
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
	// http.HandleFunc("/packet/ultimate", handler.PacketHandlerUltimate)
	// http.HandleFunc("/packet/penultimate", handler.PacketHandlerPenultimate)
	// http.HandleFunc("/packet/antepenultimate", handler.PacketHandlerAntepenultimate)
	// http.HandleFunc("/konvoy", handler.KonvoyHandler)
	// http.HandleFunc("/openshift/release", handler.OpenshiftHandlerReleasee)
	// http.HandleFunc("/about/faq", handler.FaqHandler)
	// http.HandleFunc("/nativek8s", handler.Nativek8sHandler)
	// http.HandleFunc("/delete/pipeline", handler.DeletePipeline)
	// http.HandleFunc("/os/{id:key}", GetBranch)
	r := mux.NewRouter()
	r.HandleFunc("/status", handler.StatusGitLab)
	r.HandleFunc("/{platform}/{branch}", handler.OpenshiftHandlerReleasee)
	r.HandleFunc("/{platform}/{branch}/pipeline/{id}", handler.GetPipelineDataAPI)
	r.HandleFunc("/{platform}/{branch}/job/{id}/raw", handler.GetJobLogs)

	// Trigger db update function
	go handler.UpdateDatabase()
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:3000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	glog.Infof("Listening on http://0.0.0.0:3000")
	log.Fatal(srv.ListenAndServe())
}

// func ArticlesCategoryHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "Category: %v\n", vars["key"])
// }
