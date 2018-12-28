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
}

func (s *Server) run() {
	for {
		select {
		case <-s.quit:
			s.close()
			return
		case p := <-s.register:
			golog.Info("[server.run] register peer: ", p.addr)
		case p := <-s.unregister:
			golog.Info("[server.run] unregister peer: ", p.addr)
		case <-s.ticker.C:

		case n := <-s.config.Kb.Outbox:
			switch n.Type {
			case kbs.MailPingSync:
				s.DialAndPing(n.Arg[0].(kbs.Node), true, n.Result)
			case kbs.MailFindSync:
				s.DialAndFind(n.Arg[0].(kbs.NodeID), n.Arg[1].(kbs.Node), true, n.Result)
			case kbs.MailPingAsync:
				s.DialAndPing(n.Arg[0].(kbs.Node), false, nil)
			case kbs.MailFindAsync:
				s.DialAndFind(n.Arg[0].(kbs.NodeID), n.Arg[1].(kbs.Node), false, nil)
			}
		}
	}
}

func (s *Server) handleMessage(p *Peer, m *Message) error {
	defer func(p *Peer) {
		if p.sync {
			close(p.result)
		}
	}(p)
	switch m.code {
	case MSGPing:
		var ping PingMsg
		err := json.Unmarshal(m.data, &ping)
		if err != nil {
			golog.Error("[handlePing]", err)
			return err
		}
		n := kbs.NewNode(ping.NodeID, m.ip[:], m.port)
		s.config.Kb.AddNode(n)
		return s.sendPong(p, s.config.Kb.Self.ID)
	case MSGPong:
		var pong PongMsg
		err := json.Unmarshal(m.data, &pong)
		if err != nil {
			golog.Error("[handlePong]", err)
			return err
		}
		if p.sync {
			p.result <- pong.NodeID
		}
		n := kbs.NewNode(pong.NodeID, m.ip[:], m.port)
		s.config.Kb.AddNode(n)
		p.Disconnect(errors.New("Positive"))
		return nil
	case MSGFind:
		var find FindMsg
		err := json.Unmarshal(m.data, &find)
		if err != nil {
			golog.Error("[handleMessage]", err)
			return err
		}
		golog.Info("[handleFind] ", string(m.data))
		n := kbs.NewNode(find.NodeID, m.ip[:], m.port)
		s.config.Kb.AddNode(n)
		ns, err := s.config.Kb.FindLocal(find.FindID, true)
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
		golog.Info("[handleFindAck]", findack.ToJson())
		n := kbs.NewNode(findack.NodeID, m.ip[:], m.port)
		s.config.Kb.AddNode(n)
		for _, v := range findack.Nodes {
			s.config.Kb.AddNode(v)
		}
		if p.sync {
			p.result <- findack.Nodes
		}
		p.Disconnect(errors.New("Positive"))
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

func (s *Server) DialAndPing(n kbs.Node, async bool, result chan interface{}) error {
	address := fmt.Sprintf("%s:%d", n.IP.String(), n.Port)
	p, err := s.tran.Dial(address, outtime, async, result)
	if err != nil {
		close(result)
		golog.Error("[server.dial] err: ", err)
		return err
	}
	return s.sendPing(p)
}

func (s *Server) DialAndFind(nid kbs.NodeID, n kbs.Node, async bool, result chan interface{}) error {
	address := fmt.Sprintf("%s:%d", n.IP.String(), n.Port)
	p, err := s.tran.Dial(address, outtime, async, result)
	if err != nil {
		golog.Error("[server.dial] err: ", err)
		return err
	}
	return s.sendFind(p, nid)
}

func (s *Server) Shutdown() {
	s.quit <- struct{}{}
}

func (s *Server) close() {
	s.tran.Close()
	s.ticker.Stop()
}
