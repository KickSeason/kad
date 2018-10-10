package config

import (
	"encoding/json"
	"io/ioutil"
	"kad/node"

	"github.com/kataras/golog"
)

//NodeID the id of this node
var NodeID node.NodeID

//Peers peers that configured
var Peers []string

func init() {
	data, err := ioutil.ReadFile("./node.json")
	if err != nil {
		golog.Error(err)
	}
	var config struct {
		NodeID string `json:"NodeID"`
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		golog.Error(err)
	}
	NodeID, err = node.NewIDFromString(config.NodeID)
	if err != nil {
		golog.Error(err)
		NodeID = node.NewNodeID()
		flush()
	}
}

func flush() {
	info := struct {
		NodeID string `json:"NodeID"`
	}{
		NodeID: NodeID.String(),
	}
	data, err := json.Marshal(info)
	if err != nil {
		golog.Error(err)
	}
	err = ioutil.WriteFile("./node.json", data, 0644)
	if err != nil {
		golog.Error(err)
	}
}
