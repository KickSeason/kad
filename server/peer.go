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
	timer  *time.Timer
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
