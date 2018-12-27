package kbs

import (
	"fmt"
	"net"
)

type (
	//State state
	state uint8
	//Node information about a node
	Node struct {
		Port  uint32
		ID    NodeID
		IP    net.IP
		state state
	}
)

const (
	//NSNil nil
	nsnil state = 0x01
	//NSWaitPong send ping wait for pong
	nsping  state = 0x02
	nsfound state = 0x04
)

func NewNode(nid NodeID, ip net.IP, port uint32) Node {
	return Node{
		ID:    nid,
		IP:    ip,
		Port:  port,
		state: nsnil,
	}
}

func (n *Node) ToJson() string {
	return fmt.Sprintf(`{"ID": "%s", "IP": "%s", "Port": %d}`, n.ID.String(), n.IP.String(), n.Port)
}
