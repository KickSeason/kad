package server

import (
	"net"

	"github.com/KickSeason/kad/kbs"
)

type RemoteNode struct {
	ID   kbs.NodeID
	IP   net.IP
	Port uint32
	peer *Peer
}

func NewRemoteNode(p *Peer, ip net.IP, port uint32, nid kbs.NodeID) *RemoteNode {
	return &RemoteNode{
		ID:   nid,
		IP:   ip,
		Port: port,
		peer: p,
	}
}
