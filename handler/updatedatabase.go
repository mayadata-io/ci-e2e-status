package handler

import (
	"os"
	"time"
)

// UpdateDatabase will update the latest pipelines detail and status
func UpdateDatabase() {
	token, ok := os.LookupEnv(token)
	if !ok {
		panic("TOKEN environment variable required but not set")
	}
	tick := time.Tick(60000 * time.Millisecond)
	for range tick {
		go GkeData(token)
		go EksData(token)
		go BuildData(token)
		go AwsData(token)
		go AksData(token)
		go PacketData(token)
		go GcpData(token)
	}
}
