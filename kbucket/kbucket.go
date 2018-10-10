package kbucket

import (
	"kad/node"
)

//Kbucket kbucket implement
type Kbucket struct {
	routes map[int][]node.Node
	self   node.Node
	k      int
}

const (
	k     = 8
	alpha = 3
)

//New create a kbucket
func New(nodeID node.NodeID) *Kbucket {
	return &Kbucket{}
}

//Establish init a kbucket
func (k *Kbucket) Establish() {

}

//FindNode to find a node by nodeID
func (k *Kbucket) FindNode() {

}

//AddNode to add a node
func (k *Kbucket) AddOrUpdate(node node.Node) {

}
