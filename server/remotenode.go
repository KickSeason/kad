package server

import (
	"net"

	"github.com/KickSeason/kad/kbucket"
)

type RemoteNode struct {
	ID   kbucket.NodeID
	IP   net.IP
	Port uint32
	peer *Peer
}

func NewRemoteNode(p *Peer, ip net.IP, port uint32, nid kbucket.NodeID) *RemoteNode {
	return &RemoteNode{
		ID:   nid,
		IP:   ip,
		Port: port,
		peer: p,
	}
}
