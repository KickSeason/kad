package main

import (
	"fmt"
	"kad/config"
	"kad/kbucket"
	"kad/node"
	"kad/server"

	"github.com/kataras/golog"
)

const (
	localhost = "127.0.0.1"
)

func main() {
	golog.Info("nodeid: " + string(config.NodeID.String()))
	n := &node.Node{
		Addr: fmt.Sprintf("%s:%d", localhost, config.Port),
		ID:   config.NodeID,
	}
	bucket := kbucket.New(n)
	s := server.Config{
		Addr:    fmt.Sprintf("%s:%d", localhost, config.Port),
		ID:      config.NodeID,
		Kbucket: bucket,
		Seeds:   config.Seeds,
	}
	srv := server.NewServer(s)
	srv.Start()
}
