package main

import (
	"kad/config"
	"kad/kbucket"
	"kad/server"

	"github.com/kataras/golog"
)

func main() {
	golog.Info("nodeid: " + string(config.NodeID.String()))
	// bb, e := config.NodeID.ToByte()
	// if e != nil {
	// 	golog.Fatal(e)
	// }
	// golog.Info(bb)
	// b, e := node.NewIDFromByte(bb)
	// if e != nil {
	// 	golog.Fatal(e)
	// }
	// golog.Info(b)
	bucket := kbucket.New(config.NodeID)
	s := server.Config{
		Addr:    "127.0.0.1:15200",
		ID:      config.NodeID,
		Kbucket: bucket,
	}
	srv := server.NewServer(s)
	srv.Start()
	bucket.Establish()
}
