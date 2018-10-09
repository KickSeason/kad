package util

//NodeID type nodeid
type NodeID []byte

//NewNodeID create a new node id
func NewNodeID(uuid []byte) NodeID {
	return NodeID{}
}

//String
func (n NodeID) String() string {
	return ""
}

//Equal to compare two nodeid
func (n NodeID) Equal(m NodeID) bool {
	return true
}

//Distance calculate distance between node
func (n NodeID) Distance(m NodeID) uint64 {
	return 0
}
