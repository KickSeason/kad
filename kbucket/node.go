package kbucket

import (
	"net"
)

type (
	//State state
	State string
	//Node information about a node
	Node struct {
		ID    NodeID
		IP    net.IP
		Port  uint32
		State State
	}
)

const (
	//NSNil nil
	NSNil State = ""
	//NSWaitPong send ping wait for pong
	NSWaitPong State = "waitforpong"
)

func NewNode(nid NodeID, ip net.IP, port uint32) Node {
	return Node{
		ID:    nid,
		IP:    ip,
		Port:  port,
		State: NSNil,
	}
}
