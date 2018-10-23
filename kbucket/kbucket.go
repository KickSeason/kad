package kbucket

import (
	"kad/node"
	"time"

	"github.com/kataras/golog"
)

//Kbucket kbucket implement
type (
	KQue struct {
		q      []node.Node
		k      int
		alpha  int
		bucket *Kbucket
	}
	mtype   string
	message struct {
		mtype mtype
		data  interface{}
	}
	Kbucket struct {
		routes map[int]KQue
		self   *node.Node
		k      int
		alpha  int
		ticker *time.Ticker
		pong   chan message
		Ping   chan node.Node
	}
)

const (
	mdelnode mtype = "deletenode"
	maddnode mtype = "addnode"
)
const (
	kcount = 8
	alpha  = 3
	ticktm = 5 * time.Second
)

//New create a kbucket
func New(n *node.Node) *Kbucket {
	k := &Kbucket{
		routes: make(map[int]KQue, 64),
		self:   n,
		k:      kcount,
		alpha:  alpha,
		pong:   make(chan message),
		Ping:   make(chan node.Node),
	}
	k.ticker = time.NewTicker(ticktm)
	go k.run()
	return k
}

func (kq *KQue) count() int {
	return len(kq.q)
}

func (kq *KQue) find() []node.Node {
	if kq.count() <= kq.alpha {
		res := make([]node.Node, kq.count())
		for i, v := range kq.q {
			res[i] = v
		}
		return res
	}

}

func (kq *KQue) findOne() (bool, node.Node) {

}

func (kq *KQue) has(n node.Node) bool {
	for _, v := range kq.q {
		if v.ID.Equal(n.ID) {
			return true
		}
	}
	return false
}
func (kq *KQue) updaten(n node.Node) {
	arr := []node.Node{}
	for _, v := range kq.q {
		if !v.ID.Equal(n.ID) {
			arr = append(arr, v)
		}
	}
	arr = append(arr, n)
	kq.q = arr
}
func (kq *KQue) updateAdd(n node.Node) {
	if kq.has(n) {
		kq.updaten(n)
		return
	}
	if len(kq.q) < kq.k {
		kq.q = append(kq.q, n)
		return
	}
	head := kq.q[0]
	if head.State == node.NSWaitPong {
		kq.updaten(head)
		return
	}
	if head.State == node.NSNil {
		head.State = node.NSWaitPong
		kq.q[0] = head
		kq.bucket.Ping <- head
	}
}

func (kq *KQue) remove(n node.Node) {
	if !kq.has(n) {
		return
	}
	arr := []node.Node{}
	for _, v := range kq.q {
		if !v.ID.Equal(n.ID) {
			arr = append(arr, v)
		}
	}
	kq.q = arr
}

func (k *Kbucket) newKQue() KQue {
	return KQue{
		q:      make([]node.Node, 0, k.k),
		k:      k.k,
		alpha:  k.alpha,
		bucket: k,
	}
}

func (k *Kbucket) run() {
	for {
		select {
		case <-k.ticker.C:
			golog.Info("[kbucket.run] routes: ", k.routes)
		case msg := <-k.pong:
			switch msg.mtype {
			case maddnode:
				n := msg.data.(node.Node)
				k.add(n)
			case mdelnode:
				n := msg.data.(node.Node)
				k.remove(n)
			}
		}
	}
}

//AddNode to add a node
func (k *Kbucket) AddNode(n node.Node) {
	k.pong <- message{
		mtype: maddnode,
		data:  n,
	}
}
func (k *Kbucket) add(n node.Node) {
	distance, err := node.CalDistance(n.ID, k.self.ID)
	if err != nil {
		golog.Error(err)
	}
	partion := distance.Partion()
	var que KQue
	if _, ok := k.routes[partion]; !ok {
		que = k.newKQue()
	} else {
		que = k.routes[partion]
	}
	qptr := &que
	qptr.updateAdd(n)
	k.routes[partion] = que
	return
}

//RemoveNode to remove a node
func (k *Kbucket) RemoveNode(n node.Node) {
	k.pong <- message{
		mtype: mdelnode,
		data:  n,
	}
}
func (k *Kbucket) remove(n node.Node) {
	distance, err := node.CalDistance(n.ID, k.self.ID)
	if err != nil {
		golog.Error(err)
	}
	partion := distance.Partion()
	if _, ok := k.routes[partion]; !ok {
		return
	}
	que := k.routes[partion]
	que.remove(n)
	k.routes[partion] = que
	return
}

func (k *Kbucket) Find(nid node.NodeID) []node.Node {

}

func (k *Kbucket) FindOne(nid node.NodeID) (node.Node, error) {
	distance, err := node.CalDistance(nid, k.self.ID)
	if err != nil {
		golog.Error(err)
		return node.Node{}, nil
	}
	partion := distance.Partion()
	kq, ok := k.routes[partion]
	if ok {
		if 0 < kq.count() {

		}
	}
}

func findLargestN(k int, nid node.NodeID, nodes []node.Node) (node.Node, error) {
	tmp := make([]node.Node, len(nodes))
	for i, v := range nodes {
		tmp[i] = v
	}
	pos, err := quickSort(nid, tmp, 0, len(nodes)-1)
	if err != nil {
		return node.Node{}, err
	}
	start := 0
	end := len(tmp) - 1
	for pos != k {
		if pos < k {
			start = pos + 1
		} else {
			end = pos - 1
		}
		pos, err = quickSort(nid, nodes, start, end)
		if err != nil {
			return node.Node{}, err
		}
	}
	return tmp[pos], nil
}

func quickSort(nid node.NodeID, nodes []node.Node, start int, end int) (int, error) {
	flag := nodes[start]
	dis, err := node.CalDistance(flag.ID, nid)
	if err != nil {
		golog.Error(err)
		return 0, err
	}
	i := start
	j := end
	for i < j {
		for i < j {
			d, err := node.CalDistance(nid, tmp[i].ID)
			if 0 < dis.Compare(d) {
				break
			}
		}
		for i < j {
			d, err := node.CalDistance(nid, tmp[i].ID)
			if dis.Compare(d) < 0 {
				break
			}
		}
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	return 0, nil
}
