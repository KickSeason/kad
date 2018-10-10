package server

import (
	"net"
)

//Peer a remote node
type Peer struct {
	addr string
	conn *net.TCPConn
}

//Peers remote nodes
type Peers map[string]Peer

//NewPeers create a new peers
func NewPeers() *Peers {
	p := Peers(make(map[string]Peer, 64))
	return &p
}
func (ps Peers) Contain(p Peer) bool {
	_, ok := ps[p.addr]
	return ok
}
func (ps Peers) Add(p Peer) {
	ps[p.addr] = p
}

func (ps Peers) Remove(p Peer) {
	delete(ps, p.addr)
}

//Disconnect
func (p Peer) Disconnect(err error) {

}

func (p Peer) Write(m *Message) error {
	return m.Encode(p.conn)
}
