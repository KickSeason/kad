package server

import (
	"net"
	"time"

	"github.com/kataras/golog"
)

const (
	peerOut = 10 * time.Second
)

//Peer a remote node
type Peer struct {
	addr   string
	closed bool
	once   bool
	result chan interface{}
	done   chan error
	conn   *net.TCPConn
}

func NewPeer(addr string, isonce bool, result chan interface{}, conn *net.TCPConn) *Peer {
	p := &Peer{
		addr:   conn.RemoteAddr().String(),
		closed: false,
		once:   isonce,
		result: result,
		done:   make(chan error, 1),
		conn:   conn,
	}
	return p
}

//Disconnect
func (p *Peer) Disconnect(err error) {
	golog.Info("[Peer.Disconnect] disconnect peer: ", p.addr, " reason: ", err)
	p.close(err)
}

func (p *Peer) Write(m *Message) error {
	return m.Encode(p.conn)
}

func (p *Peer) close(err error) {
	golog.Warn("[peer.close] close peer: ", p.addr)
	p.conn.Close()
}
