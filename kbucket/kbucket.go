package kbucket

//Kbucket kbucket implement
type Kbucket struct {
}

//Node information about a node
type Node struct {
	NodeID []byte
	IP     string
	Port   int
}

//New create a kbucket
func New(nodeID []byte) *Kbucket {
	return &Kbucket{}
}

//Initialize init a kbucket
func (k *Kbucket) Initialize() {

}

//FindNode to find a node by nodeID
func (k *Kbucket) FindNode() {

}

//AddNode to add a node
func (k *Kbucket) AddNode(node Node) {

}
