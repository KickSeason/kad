package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/KickSeason/kad/config"
	"github.com/KickSeason/kad/kbucket"

	"github.com/kataras/golog"
)

//Config configuration of a server
type Config struct {
	IP      net.IP
	Port    uint32
	Kbucket *kbucket.Kbucket
	Seeds   []string
}

//Server a tcp server
type Server struct {
	config     Config
	tran       *Transport
	peers      map[string]*Peer
	remotes    map[string]*RemoteNode
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
		remotes:    make(map[string]*RemoteNode, 64),
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
			golog.Info("[server.tick] ", s.remotes)
		case n := <-s.config.Kbucket.Sender:
			switch n.Type {
			case kbucket.NPing:
				s.dial(n.Arg.(kbucket.Node))
			}
		}
	}
}

func (s *Server) handleMessage(p *Peer, m *Message) error {
	switch m.code {
	case MSGPing:
		var ping PingMsg
		err := json.Unmarshal(m.data, ping)
		if err != nil {
			golog.Error(err)
			return err
		}
		n := NewRemoteNode(p, m.ip[:], m.port, ping.NodeID)
		s.addRemoteNode(n)
		return s.sendPong(p, s.config.Kbucket.Self.ID)
	case MSGPong:
		var pong PongMsg
		err := json.Unmarshal(m.data, pong)
		if err != nil {
			golog.Error(err)
			return err
		}
		n := kbucket.Node{
			ID:   pong.NodeID,
			IP:   m.ip[:],
			Port: m.port,
		}
		rn := NewRemoteNode(p, m.ip[:], m.port, pong.NodeID)
		s.addRemoteNode(rn)
		s.config.Kbucket.AddNode(n)
		return nil
	case MSGFind:
		var find FindMsg
		err := json.Unmarshal(m.data, find)
		if err != nil {
			golog.Error(err)
			return err
		}
		ns, err := s.config.Kbucket.Find(find.NodeID)
		if err != nil {
			golog.Error(err)
			return err
		}
		return s.sendFindAck(p, find.NodeID, ns)
	case MSgFindAck:
	}
	return nil
}
func (s *Server) sendFind(p *Peer, nid kbucket.NodeID) error {
	find := FindMsg{
		NodeID: nid,
	}
	data, err := json.Marshal(find)
	if err != nil {
		return err
	}
	msg := NewMessage(MAGIC, MSGFind, s.config.IP, s.config.Port, data)
	return p.Write(msg)
}
func (s *Server) sendFindAck(p *Peer, nid kbucket.NodeID, ns []kbucket.Node) error {
	findack := FindAckMsg{
		NodeID: nid,
		Nodes:  ns,
	}
	data, err := json.Marshal(findack)
	if err != nil {
		return err
	}
	msg := NewMessage(MAGIC, MSGFind, s.config.IP, s.config.Port, data)
	return p.Write(msg)
}

func (s *Server) sendPong(p *Peer, nid kbucket.NodeID) error {
	idbyte, err := nid.ToByte()
	if err != nil {
		return err
	}
	msg := NewMessage(MAGIC, MSGPong, s.config.IP, s.config.Port, idbyte)
	return p.Write(msg)
}
func (s *Server) sendPing(p *Peer) error {
	id, err := config.NodeID.ToByte()
	if err != nil {
		return err
	}
	msg := NewMessage(MAGIC, MSGPing, s.config.IP, s.config.Port, id)
	return p.Write(msg)
}

func (s *Server) addPeer(p *Peer) {
	if _, ok := s.peers[p.addr]; ok {
		return
	}
	s.peers[p.addr] = p
}
func (s *Server) removePeer(p *Peer) {
	delete(s.peers, p.addr)
	key := ""
	for k, v := range s.remotes {
		if v.peer == p {
			key = k
			break
		}
	}
	delete(s.remotes, key)
}

func (s *Server) addRemoteNode(rn *RemoteNode) {
	address := fmt.Sprintf("%s:%d", rn.IP, rn.Port)
	s.remotes[address] = rn
}
func (s *Server) dial(n kbucket.Node) error {
	address := fmt.Sprintf("%s:%d", n.IP.String(), n.Port)
	if p, ok := s.peers[address]; ok {
		return s.sendPing(p)
	}
	p, err := s.tran.Dial(address, outtime)
	if err != nil {
		golog.Error("[server.dial] err: ", err)
		return err
	}
	return s.sendPing(p)
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
