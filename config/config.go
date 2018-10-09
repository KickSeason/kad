package config

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/google/uuid"
)

//NodeID the id of this node
var NodeID []byte

func init() {
	test()
	data, err := ioutil.ReadFile("./config.json")
	defer func() {
		if len(NodeID) == 0 {
			NodeID = uuid.NodeID()
		}
		flush()
	}()
	if err != nil {
		log.Println(err)
	}
	var config struct {
		NodeID []byte `json:"NodeID"`
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Println(err)
	}
	NodeID = config.NodeID
}

func flush() {
	info := struct {
		NodeID []byte `json:"NodeID"`
	}{
		NodeID: NodeID,
	}
	data, err := json.Marshal(info)
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile("./config.json", data, 0644)
	if err != nil {
		log.Println(err)
	}
}

func test() {
	n := []byte{0, 0}
	for _, v := range uuid.NodeID() {
		n = append(n, byte(v))
	}
	var mySlice = []byte{244, 244, 244, 244, 244, 244, 244, 244}
	data := binary.BigEndian.Uint64(mySlice)
	fmt.Println(data)
	fmt.Println(n)
	fmt.Println(mySlice)
	data = binary.BigEndian.uin
	fmt.Println(data)
	uid := uuid.New()
	fmt.Println(uid.String())
	fmt.Println(uid.ID())
}
