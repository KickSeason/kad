package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/KickSeason/kad/config"
	"github.com/KickSeason/kad/kbs"
	"github.com/KickSeason/kad/server"

	"github.com/c-bata/go-prompt"
)

var running = true

func interactor(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "table", Description: "get hash table"},
		{Text: "store", Description: "store key-value into network"},
		{Text: "exit", Description: "exit"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
func main() {
	fmt.Println("nodeid: " + string(config.NodeID.String()))
	ip := net.ParseIP(config.Address)
	c := &kbs.KbConfig{
		Seeds:   config.Seeds,
		LocalIP: ip,
		Port:    config.Port,
		ID:      config.NodeID,
	}
	kb := kbs.NewKBS(c)
	s := server.Config{
		IP:    ip,
		Port:  config.Port,
		Kb:    kb,
		Seeds: config.Seeds,
	}
	srv := server.NewServer(s)
	srv.Start()
	kb.Start()
	fmt.Println("my kademlia impl")
	go kb.Find(config.NodeID)
	for running {
		t := prompt.Input("> ", interactor)
		switch t {
		case "exit":
			running = false
		case "table":
			fmt.Println(kb.ToJson())
		default:
			inputs := strings.Split(t, " ")
			if inputs[0] == "store" {
				if 3 == len(inputs) && inputs[1] != "" && inputs[2] != "" {
					kb.Store(inputs[1], inputs[2])
				}
			} else if inputs[0] == "get" {
				if 2 == len(inputs) && inputs[1] != "" {
					value, err := kb.Get(inputs[1])
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(value)
					}
				}
			}
		}
	}
}
