package server

import (
	"errors"
	"kad/config"
	"kad/kbucket"
	"kad/node"
	"time"

	"github.com/kataras/golog"
)

//Config configuration of a server
type Config struct {
	Addr    string
	ID      node.NodeID
	Kbucket *kbucket.Kbucket
	Seeds   []string
}

//Server a tcp server
type Server struct {
	config     Config
	tran       *Transport
	peers      map[string]*Peer
	register   chan *Peer
	unregister chan *Peer
	quit       chan struct{}
	errch      chan error
	ticker     *time.Ticker
}

//NewServer to create a new server
func NewServer(config Config) *Server {
	s := &Server{
		config:     config,
		peers:      make(map[string]*Peer, 64),
		register:   make(chan *Peer),
		unregister: make(chan *Peer),
		errch:      make(chan error),
	}
	s.tran = NewTransport(s)
	s.ticker = time.NewTicker(3 * time.Second)
	return s
}

const (
	outtime = 5 * time.Second
)

//Start to start a server
func (s *Server) Start() {
	go func() {
		for _, v := range s.config.Seeds {
			p, err := s.tran.Dial(v, outtime)
			if err != nil {
				golog.Error("[server.tran.dia]", err)
				continue
			}
			s.sendPing(p)
		}
	}()
	go s.tran.Accept()
	s.run()
}

func (s *Server) run() {
	for {
		select {
		case <-s.quit:
			s.close()
			return
		case p := <-s.register:
			s.addPeer(p)
		case p := <-s.unregister:
			golog.Info("[server.run] unregister peer: ", p)
			s.removePeer(p)
		case <-s.ticker.C:
			golog.Info("[server.tick] ", s.peers)
		case n := <-s.config.Kbucket.Ping:
			err := s.dial(n)
			if err != nil {
				s.config.Kbucket.RemoveNode(n)
			}
		}
	}
}

func (s *Server) handleMessage(p *Peer, m *Message) error {
	if m.code == CodeMap[MSGPing] {
		return s.handlePing(p, m)
	}
	if m.code == CodeMap[MSGPong] {
		return s.handlePong(p, m)
	}
	return nil
}

func (s *Server) handlePing(p *Peer, m *Message) error {
	nid, err := node.NewIDFromByte(m.data)
	if err != nil {
		golog.Error("[handleping] ", err)
		return err
	}
	p.SetID(nid)
	n := node.Node{
		ID:   nid,
		Addr: p.addr,
	}
	s.config.Kbucket.AddNode(n)
	golog.Info("[server.handleping] recieve ping from node: ", p.addr, nid.String())
	return s.sendPong(p, s.config.ID)
}

func (s *Server) handlePong(p *Peer, m *Message) error {
	nid, err := node.NewIDFromByte(m.data)
	if err != nil {
		return err
	}
	p.SetID(nid)
	n := node.Node{
		ID:   nid,
		Addr: p.addr,
	}
	golog.Info("[server.handlepong] recieve pong from node: ", p.addr, nid.String())
	s.config.Kbucket.AddNode(n)
	return nil
}

func (s *Server) sendPong(p *Peer, nid node.NodeID) error {
	idbyte, err := nid.ToByte()
	if err != nil {
		return err
	}
	msg := NewMessage(MAGIC, MSGPong, idbyte)
	return p.Write(msg)
}
func (s *Server) sendPing(p *Peer) error {
	id, err := config.NodeID.ToByte()
	if err != nil {
		return err
	}
	msg := NewMessage(MAGIC, MSGPing, id)
	return p.Write(msg)
}
func (s *Server) addPeer(p *Peer) {
	if _, ok := s.peers[p.addr]; ok {
		return
	}
	s.peers[p.addr] = p
}
func (s *Server) dial(n node.Node) error {
	if p, ok := s.peers[n.Addr]; ok {
		return s.sendPing(p)
	}
	p, err := s.tran.Dial(n.Addr, outtime)
	if err != nil {
		golog.Error("[server.dial] err: ", err)
		return err
	}
	return s.sendPing(p)
}

func (s *Server) removePeer(p *Peer) {
	delete(s.peers, p.addr)
	n := node.Node{
		Addr: p.addr,
		ID:   p.id,
	}
	s.config.Kbucket.RemoveNode(n)
}

func (s *Server) Shutdown() {
	s.quit <- struct{}{}
}

func (s *Server) close() {
	s.tran.Close()
	for _, v := range s.peers {
		v.Disconnect(errors.New("Stopped manully"))
	}
	s.ticker.Stop()
}
