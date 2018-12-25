package main

import (
	"net"

	"github.com/KickSeason/kad/config"
	"github.com/KickSeason/kad/kbucket"
	"github.com/KickSeason/kad/server"

	"github.com/c-bata/go-prompt"
	"github.com/kataras/golog"
)

var running = true

func interactor(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "htable", Description: "get hash table"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
func main() {
	golog.Info("nodeid: " + string(config.NodeID.String()))
	ip := net.ParseIP(config.Address)
	c := &kbucket.KbConfig{
		Seeds:   config.Seeds,
		LocalIP: ip,
		Port:    config.Port,
		ID:      config.NodeID,
	}
	kb := kbucket.New(c)
	s := server.Config{
		IP:      ip,
		Port:    config.Port,
		Kbucket: kb,
		Seeds:   config.Seeds,
	}
	srv := server.NewServer(s)
	srv.Start()
	kb.Start()
	golog.Info("my kademlia impl")
	for running {
		t := prompt.Input("> ", interactor)
		golog.Info("selected: ", t)
		switch t {
		case "exit":
			running = false
		case "htable":

		}
	}
}
