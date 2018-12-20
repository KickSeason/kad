package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	MAGIC = 0x7596
)

type Message struct {
	magic  uint32
	code   uint32
	length uint32
	data   []byte
}
type MessageType string

const (
	MSGPing    MessageType = "ping"
	MSGPong    MessageType = "pong"
	MSGFind    MessageType = "find"
	MSgFindAck MessageType = "fdack"
)

var CodeMap = map[MessageType]uint32{
	MSGPing: 0x00F1,
	MSGPong: 0x01F1,
}

func NewMessage(magic uint32, mtype MessageType, data []byte) *Message {
	return &Message{
		magic:  magic,
		code:   CodeMap[mtype],
		length: uint32(len(data)),
		data:   data,
	}
}

func (m *Message) Encode(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &m.magic); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &m.code); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &m.length); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &m.data); err != nil {
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
