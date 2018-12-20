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
var Seeds []string

//Port server port
var Port int

//Address bind
var Address string

func init() {
	err := readConfig()
	if err != nil {
		golog.Fatal(err)
	}
	err = readNodeInfo()
	if err != nil {
		golog.Fatal(err)
	}
}
func readConfig() error {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return err
	}
	var config struct {
		Port    int      `json:"port"`
		Seeds   []string `json:"seeds"`
		Address string   `json:"address"`
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		golog.Error(err)
	}
	Port = config.Port
	Seeds = config.Seeds
	Address = config.Address
	return nil
}

func readNodeInfo() error {
	data, err := ioutil.ReadFile("./node.json")
	if err != nil {
		return err
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
		persist()
	}
	return nil
}
func persist() {
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
