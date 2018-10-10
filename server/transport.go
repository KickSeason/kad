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

func (t *Transport) Dial(addr string, ot time.Duration) error {
	var tcpaddr net.TCPAddr
	localaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", localaddr, &tcpaddr)
	if err != nil {
		return err
	}
	go t.handleConn(conn)
	return nil
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
	t.listener = listen

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			golog.Error(err)
			if t.isCloseError(err) {
				break
			}
			continue
		}
		conn.SetKeepAlive(true)
		go t.handleConn(conn)
	}
}

func (t *Transport) handleConn(conn *net.TCPConn) {
	var err error
	p := Peer{
		addr: conn.RemoteAddr().String(),
		conn: conn,
	}
	defer func() {
		p.Disconnect(err)
	}()
	t.server.register <- p
	for {
		msg := &Message{}
		if err := msg.Decode(conn); err != nil {
			golog.Error(err)
			return
		}
		if err := t.server.handleMessage(p, msg); err != nil {
			golog.Error(err)
			return
		}
	}
}

func (t *Transport) isCloseError(err error) bool {
	golog.Error(err)
	return false
}
