package server

import (
	"kad/kbucket"
	"kad/node"

	"github.com/kataras/golog"
)

//Config configuration of a server
type Config struct {
	Addr    string
	ID      node.NodeID
	Kbucket *kbucket.Kbucket
}

//Server a tcp server
type Server struct {
	config     Config
	tran       *Transport
	peers      *Peers
	register   chan Peer
	unregister chan Peer
	errch      chan error
}

//NewServer to create a new server
func NewServer(config Config) *Server {
	s := &Server{
		config:     config,
		peers:      NewPeers(),
		register:   make(chan Peer),
		unregister: make(chan Peer),
		errch:      make(chan error),
	}
	s.tran = NewTransport(s)
	return s
}

//Start to start a server
func (s *Server) Start() {
	go s.tran.Accept()
	s.run()
}

func (s *Server) run() {
	for {
		select {
		case p := <-s.register:
			s.peers.Add(p)
		case p := <-s.unregister:
			s.peers.Remove(p)
		}
	}
}

func (s *Server) handleMessage(p Peer, m *Message) error {
	if m.code == CodeMap[MSGPing] {
		return s.handlePing(p, m)
	}
	if m.code == CodeMap[MSGPong] {
		return s.handlePong(p, m)
	}
	return nil
}

func (s *Server) handlePing(p Peer, m *Message) error {
	nid, err := node.NewIDFromByte(m.data)
	if err != nil {
		golog.Error("[handleping] ", err)
		return err
	}
	n := node.Node{
		ID:   nid,
		Addr: p.addr,
	}
	s.config.Kbucket.AddOrUpdate(n)
	return s.sendPong(p, s.config.ID)
}

func (s *Server) handlePong(p Peer, m *Message) error {
	nid, err := node.NewIDFromByte(m.data)
	if err != nil {
		return err
	}
	n := node.Node{
		ID:   nid,
		Addr: p.addr,
	}
	s.config.Kbucket.AddOrUpdate(n)
	return nil
}

func (s *Server) sendPong(p Peer, nid node.NodeID) error {
	idbyte, err := nid.ToByte()
	if err != nil {
		return err
	}
	msg := NewMessage(MAGIC, MSGPong, idbyte)
	return p.Write(msg)
}
