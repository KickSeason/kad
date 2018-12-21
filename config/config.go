package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/KickSeason/kad/kbucket"
	"github.com/kataras/golog"
)

//NodeID the id of this node
var NodeID kbucket.NodeID

//Seeds peers that configured
var Seeds []string

//Port server port
var Port uint32

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
		Port    uint32   `json:"port"`
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
	NodeID, err = kbucket.NewIDFromString(config.NodeID)
	if err != nil {
		golog.Error(err)
		NodeID = kbucket.NewNodeID()
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
