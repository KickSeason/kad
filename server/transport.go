package server

import (
	"net"
	"time"

	"github.com/kataras/golog"
)

type Transport struct {
	server   *Server
	listener *net.TCPListener
}

func NewTransport(s *Server) *Transport {
	return &Transport{
		server: s,
	}
}

func (t *Transport) Dial(addr string, ot time.Duration) (*Peer, error) {
	destaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return &Peer{}, err
	}
	conn, err := net.DialTCP("tcp", nil, destaddr)
	if err != nil {
		return &Peer{}, err
	}
	p := NewPeer(conn.RemoteAddr().String(), conn)
	go t.handleConn(p)
	return p, nil
}

func (t *Transport) Accept() {
	hawServer, err := net.ResolveTCPAddr("tcp", t.server.config.Addr)
	if err != nil {
		golog.Fatal(err)
	}
	listen, err := net.ListenTCP("tcp", hawServer)
	if err != nil {
		golog.Fatal(err)
	}
	golog.Info("[transport.accept] listening on: ", hawServer.String())
	t.listener = listen

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			golog.Error("[transport][accept]", err)
			if t.isCloseError(err) {
				break
			}
			continue
		}
		conn.SetKeepAlive(true)
		p := NewPeer(conn.RemoteAddr().String(), conn)
		go t.handleConn(p)
	}
}

func (t *Transport) handleConn(p *Peer) {
	var err error
	defer func() {
		p.Disconnect(err)
		t.server.unregister <- p
	}()
	t.server.register <- p
	for {
		msg := &Message{}
		if err = msg.Decode(p.conn); err != nil {
			golog.Error(err)
			return
		}
		if err = t.server.handleMessage(p, msg); err != nil {
			golog.Error(err)
			return
		}
	}
}

func (t *Transport) isCloseError(err error) bool {
	golog.Error(err)
	return false
}

func (t *Transport) Close() {
	t.listener.Close()
}
