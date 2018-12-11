package main

import (
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/openebs/ci-e2e-dashboard-go-backend/database"
	"github.com/openebs/ci-e2e-dashboard-go-backend/handler"
	log "github.com/sirupsen/logrus"
)

func main() {
	database.InitDb()
	http.HandleFunc("/gke", handler.Gkehandler)
	http.HandleFunc("/aws", handler.Awshandler)
	http.HandleFunc("/gcp", handler.Gcphandler)
	http.HandleFunc("/aks", handler.Akshandler)
	http.HandleFunc("/packet", handler.Packethandler)
	http.HandleFunc("/eks", handler.Ekshandler)
	http.HandleFunc("/build", handler.Buildhandler)
	fmt.Println("Listening on http://localhost:3000")
	log.Fatal(http.ListenAndServe("localhost:3000", nil))
}
