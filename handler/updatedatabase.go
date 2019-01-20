package handler

import (
	"fmt"
	"os"

	"github.com/golang/glog"
)

// UpdateDatabase will update the latest pipelines detail and status
func UpdateDatabase() {
	// Read token environment variable
	token, ok := os.LookupEnv(token)
	if !ok {
		glog.Fatalf("TOKEN environment variable required")
	}
	BuildData(token)
	// GkeData(token)
	fmt.Println("finish function")
	// loop will iterate at every 6000 seconds
	// tick := time.Tick(60000 * time.Millisecond)
	// for range tick {
	// 	go BuildData(token)
	// 	fmt.Println("function finish")

	// // Trigger GkeData function for update GKE related data to database
	// go GkeData(token)
	// // Trigger GkeData function for update EKS related data to database
	// go EksData(token)
	// // Trigger GkeData function for update Build related data to database
	// // Trigger GkeData function for update AWS related data to database
	// go AwsData(token)
	// // Trigger GkeData function for update AKS related data to database
	// go AksData(token)
	// // Trigger GkeData function for update PACKET related data to database
	// go PacketData(token)
	// // Trigger GkeData function for update GCP related data to database
	// go GcpData(token)
	// }
}
