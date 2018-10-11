package node

type (
	//State state
	State string
	//Node information about a node
	Node struct {
		ID    NodeID
		Addr  string
		State State
	}
)

const (
	//NSNil nil
	NSNil State = ""
	//NSWaitPong send ping wait for pong
	NSWaitPong State = "waitforpong"
)
