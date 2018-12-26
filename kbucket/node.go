package kbs

import (
	"net"
)

type (
	//State state
	state uint8
	//Node information about a node
	Node struct {
		Port uint32
		ID   NodeID
		IP   net.IP
	}
)

const (
	//NSNil nil
	nsnil state = 0x01
	//NSWaitPong send ping wait for pong
	nsping state = 0x02
)

func NewNode(nid NodeID, ip net.IP, port uint32) Node {
	return Node{
		ID:   nid,
		IP:   ip,
		Port: port,
	}
}
