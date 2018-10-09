package main

import (
	"kad/config"
	"kad/kbucket"
	"kad/server"
	"log"
)

func main() {
	log.Println("nodeid: " + string(config.NodeID))
	srv := server.NewServer(config.NodeID)
	srv.Start()
	bucket := kbucket.New(config.NodeID)
	bucket.Initialize()
}
