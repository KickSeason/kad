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
	conn   *net.TCPConn
	timer  *time.Timer
	done   chan error
	closed bool
}

func NewPeer(addr string, conn *net.TCPConn) *Peer {
	p := &Peer{
		addr:   conn.RemoteAddr().String(),
		conn:   conn,
		done:   make(chan error, 1),
		closed: false,
	}
	p.timer = time.NewTimer(peerOut)
	go p.run()
	return p
}

func (p *Peer) run() {
	for {
		select {
		case <-p.timer.C:
			// err := p.server.sendPing(p)
			// if err != nil {
			// 	golog.Error(err)
			// }
			continue
		}
	}
}

//Disconnect
func (p *Peer) Disconnect(err error) {
	golog.Info("[Peer.Disconnect] disconnect peer: ", p.addr, " reason: ", err)
	p.close(err)
}

func (p *Peer) Write(m *Message) error {
	p.timer.Reset(peerOut)
	return m.Encode(p.conn)
}

func (p *Peer) close(err error) {
	golog.Warn("[peer.close] close peer: ", p.addr)
	p.timer.Stop()
	p.conn.Close()
}
