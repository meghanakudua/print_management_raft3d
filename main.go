// main.go
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/meghanakudua/print_management_raft3d/api"
	"github.com/meghanakudua/print_management_raft3d/raft"
)

func main() {
	nodeID := flag.String("id", "node1", "Unique Raft node ID")
	bindAddr := flag.String("bind", ":9001", "Bind address for Raft communication")
	dataDir := flag.String("data", "data", "Raft data directory")
	peersStr := flag.String("peers", "", "Comma-separated list of peer addresses (e.g., :9001,:9002,:9003)")
	bootstrap := flag.Bool("bootstrap", false, "Is this the bootstrap node?")
	apiPort := flag.String("api", "8001", "API port to listen on")

	flag.Parse()

	//peers := strings.Split(*peersStr, ",")
	var peers []string
	if *peersStr != "" {
		peers = strings.Split(*peersStr, ",")
	}

	// Create data directory for Raft if it doesn't exist
	os.MkdirAll(*dataDir, 0700)

	raftNode, err := raft.StartRaftNode(*nodeID, *bindAddr, *dataDir, peers, *bootstrap)
	if err != nil {
		log.Fatalf("Failed to start Raft node: %v", err)
	}

	log.Printf("HTTP API running at :%s...", *apiPort)
	router := api.NewRouter(raftNode)
	http.ListenAndServe(":"+*apiPort, router)
}
