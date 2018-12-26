package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/KickSeason/kad/kbs"
)

type MsgType uint8

const (
	MSGPing    MsgType = 0x01
	MSGPong    MsgType = 0x02
	MSGFind    MsgType = 0x03
	MSgFindAck MsgType = 0x04
	MSGStore   MsgType = 0x05
)

type PingMsg struct {
	NodeID kbs.NodeID `json: "nodeid"`
}

type PongMsg struct {
	NodeID kbs.NodeID `json: "nodeid"`
}

type FindMsg struct {
	NodeID kbs.NodeID `json: "nodeid"`
	FindID kbs.NodeID `json: "findid"`
}

type FindAckMsg struct {
	NodeID kbs.NodeID `json: "nodeid"`
	FindID kbs.NodeID `json: "findid"`
	Nodes  []kbs.Node `json: "nodes"`
}
type StoreMsg struct {
	NodeID kbs.NodeID `json: "nodeid"`
	key    string     `json: "key"`
	value  string     `json: "value"`
}

const (
	MAGIC uint16 = 0x7596
)

type Message struct {
	magic  uint16
	code   MsgType
	ip     [4]byte
	port   uint32
	length uint32
	data   []byte
}

func NewMessage(magic uint16, mtype MsgType, ip net.IP, port uint32, data []byte) *Message {
	m := &Message{
		magic:  magic,
		code:   mtype,
		port:   port,
		length: uint32(len(data)),
		data:   data,
	}
	iparray := []byte(ip.To4())
	for i := 0; i < 4; i++ {
		m.ip[i] = iparray[i]
	}
	return m
}

func (m *Message) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, m.magic); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, m.code); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, m.ip); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, m.port); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, m.length); err != nil {
		return err
	}
	if m.length == 0 {
		return nil
	}
	if err := binary.Write(w, binary.LittleEndian, m.data); err != nil {
		return err
	}
	return nil
}
func (m *Message) Decode(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &m.magic); err != nil {
		return err
	}
	if m.magic != MAGIC {
		return errors.New("magic not match")
	}
	if err := binary.Read(r, binary.LittleEndian, &m.code); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &m.ip); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &m.port); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &m.length); err != nil {
		return err
	}
	return m.DecodeData(r)
}

func (m *Message) DecodeData(r io.Reader) error {
	buf := new(bytes.Buffer)
	n, err := io.CopyN(buf, r, int64(m.length))
	if err != nil {
		return err
	}
	if n != int64(m.length) {
		return fmt.Errorf("expected to read %d bytes, but got %d bytes", m.length, n)
	}
	m.data = buf.Bytes()
	return nil
}
