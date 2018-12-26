package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/KickSeason/kad/kbs"

	"github.com/kataras/golog"
)

//Config configuration of a server
type Config struct {
	IP    net.IP
	Port  uint32
	Kb    *kbs.KBS
	Seeds []string
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
	go s.tran.Accept()
	go s.run()
	go func() {
		for _, v := range s.config.Seeds {
			p, err := s.tran.Dial(v, outtime, true, nil)
			if err != nil {
				golog.Error("[server.tran.dia]", err)
				continue
			}
			s.sendFind(p, s.config.Kb.Self.ID)
		}
	}()
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

		case n := <-s.config.Kb.Outbox:
			switch n.Type {
			case kbs.MailPingSync:
				s.DialAndPing(n.Arg[0].(kbs.Node), n.Result)
			case kbs.MailFindSync:
				s.DialAndFind(n.Arg[0].(kbs.NodeID), n.Arg[1].(kbs.Node))
			case kbs.MailPingAsync:
				s.DialAndPing(n.Arg[0].(kbs.Node), nil)
			case kbs.MailFindAsync:
				s.DialAndFind(n.Arg[0].(kbs.NodeID), n.Arg[1].(kbs.Node))
			}
		}
	}
}

func (s *Server) handleMessage(p *Peer, m *Message) error {
	switch m.code {
	case MSGPing:
		var ping PingMsg
		err := json.Unmarshal(m.data, &ping)
		if err != nil {
			golog.Error("[handlePing]", err)
			return err
		}
		n := kbs.NewNode(ping.NodeID, m.ip[:], m.port)
		s.addRemoteNode(p, &n)
		return s.sendPong(p, s.config.Kb.Self.ID)
	case MSGPong:
		var pong PongMsg
		err := json.Unmarshal(m.data, &pong)
		if err != nil {
			golog.Error("[handlePong]", err)
			return err
		}
		if p.once && p.result != nil {
			p.result <- pong.NodeID
			return nil
		}
		n := kbs.NewNode(pong.NodeID, m.ip[:], m.port)
		s.config.Kb.AddNode(n)
		s.addRemoteNode(p, &n)
		return nil
	case MSGFind:
		var find FindMsg
		err := json.Unmarshal(m.data, &find)
		if err != nil {
			golog.Error("[handleMessage]", err)
			return err
		}
		n := kbs.NewNode(find.NodeID, m.ip[:], m.port)
		s.config.Kb.AddNode(n)
		s.addRemoteNode(p, &n)
		ns, err := s.config.Kb.Find(find.FindID)
		if err != nil {
			golog.Error("[handleFind]", err)
			return err
		}
		return s.sendFindAck(p, find.NodeID, ns)
	case MSgFindAck:
		var findack FindAckMsg
		err := json.Unmarshal(m.data, &findack)
		if err != nil {
			golog.Error("[handleFindAck]", err)
			return err
		}
		golog.Info("findack", findack)
		n := kbs.NewNode(findack.NodeID, m.ip[:], m.port)
		s.config.Kb.AddNode(n)
		s.addRemoteNode(p, &n)
		for _, v := range findack.Nodes {
			s.config.Kb.AddNode(v)
		}
	}
	return nil
}
func (s *Server) sendFind(p *Peer, nid kbs.NodeID) error {
	find := FindMsg{
		NodeID: s.config.Kb.Self.ID,
		FindID: nid,
	}
	data, err := json.Marshal(find)
	if err != nil {
		golog.Error("[sendFind]", err)
		return err
	}
	msg := NewMessage(MAGIC, MSGFind, s.config.IP, s.config.Port, data)
	return p.Write(msg)
}
func (s *Server) sendFindAck(p *Peer, nid kbs.NodeID, ns []kbs.Node) error {
	findack := FindAckMsg{
		NodeID: s.config.Kb.Self.ID,
		FindID: nid,
		Nodes:  ns,
	}
	data, err := json.Marshal(findack)
	if err != nil {
		golog.Error("[sendFindAck]", err)
		return err
	}
	msg := NewMessage(MAGIC, MSgFindAck, s.config.IP, s.config.Port, data)
	return p.Write(msg)
}

func (s *Server) sendPong(p *Peer, nid kbs.NodeID) error {
	pong := PongMsg{
		NodeID: s.config.Kb.Self.ID,
	}
	data, err := json.Marshal(pong)
	if err != nil {
		golog.Error("[sendPong]", err)
		return err
	}
	msg := NewMessage(MAGIC, MSGPong, s.config.IP, s.config.Port, data)
	return p.Write(msg)
}
func (s *Server) sendPing(p *Peer) error {
	ping := PingMsg{
		NodeID: s.config.Kb.Self.ID,
	}
	data, err := json.Marshal(ping)
	if err != nil {
		golog.Error("[sendPing]", err)
		return err
	}
	msg := NewMessage(MAGIC, MSGPing, s.config.IP, s.config.Port, data)
	return p.Write(msg)
}

func (s *Server) DialAndPing(n kbs.Node, result chan interface{}) error {
	address := fmt.Sprintf("%s:%d", n.IP.String(), n.Port)
	if p, ok := s.peers[address]; ok {
		p.result = result
		return s.sendPing(p)
	}
	p, err := s.tran.Dial(address, outtime, true, result)
	if err != nil {
		close(result)
		golog.Error("[server.dial] err: ", err)
		return err
	}
	return s.sendPing(p)
}

func (s *Server) DialAndFind(nid kbs.NodeID, n kbs.Node) error {
	address := fmt.Sprintf("%s:%d", n.IP.String(), n.Port)
	if p, ok := s.peers[address]; ok {
		return s.sendFind(p, nid)
	}
	p, err := s.tran.Dial(address, outtime, true, nil)
	if err != nil {
		golog.Error("[server.dial] err: ", err)
		return err
	}
	return s.sendFind(p, nid)
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

func (s *Server) addRemoteNode(p *Peer, n *kbs.Node) {
	if p.once {
		p.Disconnect(errors.New("positive"))
		return
	}
	rn := NewRemoteNode(p, n.IP, n.Port, n.ID)
	address := fmt.Sprintf("%s:%d", rn.IP, rn.Port)
	s.remotes[address] = rn
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
