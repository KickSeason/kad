package main

import (
	"net"

	"github.com/KickSeason/kad/config"
	"github.com/KickSeason/kad/kbucket"
	"github.com/KickSeason/kad/server"

	"github.com/kataras/golog"
)

func main() {
	golog.Info("nodeid: " + string(config.NodeID.String()))
	ip := net.ParseIP(config.Address)
	n := &kbucket.Node{
		IP:   ip,
		Port: config.Port,
		ID:   config.NodeID,
	}
	kb := kbucket.New(n)
	s := server.Config{
		IP:      ip,
		Port:    config.Port,
		Kbucket: kb,
		Seeds:   config.Seeds,
	}
	srv := server.NewServer(s)
	srv.Start()
}
