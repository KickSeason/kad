package kbs

import (
	"github.com/google/uuid"
)

//NodeID type nodeid
type NodeID uuid.UUID

//NewNodeID create a new node id
func NewNodeID() NodeID {
	return NodeID(uuid.New())
}

//NewIDFromString create a new nodeid from string
func NewIDFromString(str string) (nid NodeID, err error) {
	uid, err := uuid.Parse(str)
	return NodeID(uid), err
}

//NewIDFromByte create a new nodeid from byte
func NewIDFromByte(b []byte) (nid NodeID, err error) {
	uid, err := uuid.FromBytes(b)
	return NodeID(uid), err
}

//String
func (n NodeID) String() string {
	return uuid.UUID(n).String()
}

//Equal to compare two nodeid
func (n NodeID) Equal(m NodeID) bool {
	return n.String() == m.String()
}

//ToByte convert into []byte
func (n NodeID) ToByte() ([]byte, error) {
	return uuid.UUID(n).MarshalBinary()
}

//Length the length of nodeid
func Length() int {
	return 128
}
