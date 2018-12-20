package kbucket

import (
	"errors"
	"kad/node"
	"sort"
	"time"

	"github.com/kataras/golog"
)

//Kbucket kbucket implement
type (
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
func New(local *node.Node) *Kbucket {
	k := &Kbucket{
		routes: make(map[int]KQue, 64),
		self:   local,
		k:      kcount,
		alpha:  alpha,
		pong:   make(chan message),
		Ping:   make(chan node.Node),
	}
	k.ticker = time.NewTicker(ticktm)
	go k.run()
	return k
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
		que = newKQue(k)
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

//Find find alpha nodes that are closest to the nid
func (k *Kbucket) Find(nid node.NodeID) (ns []node.Node, err error) {
	if k.self.ID.Equal(nid) {
		return []node.Node{*k.self}, nil
	}
	dist, err := node.CalDistance(nid, k.self.ID)
	if err != nil {
		golog.Error(err)
		return []node.Node{}, err
	}
	partion := dist.Partion()
	if kq, ok := k.routes[partion]; ok {
		ns, err := kq.findN(nid, k.alpha)
		if err != nil {
			return ns, err
		}
	}
	if k.alpha <= len(ns) {
		return ns, nil
	}
	pslice := make([]int, len(k.routes))
	i := 0
	for key := range k.routes {
		pslice[i] = key
		i++
	}
	p := Partions{
		parts: pslice,
		base:  partion,
	}
	sort.Sort(p)
	for _, v := range p.parts {
		kq, ok := k.routes[v]
		if ok {
			res, err := kq.findN(nid, k.alpha-len(ns))
			if err != nil {
				return ns, err
			}
			for _, v := range res {
				ns = append(ns, v)
			}
			if k.alpha <= len(ns) {
				break
			}
		}
	}
	return ns, nil
}

func (k *Kbucket) FindOne(nid node.NodeID) (node.Node, error) {
	if k.self.ID.Equal(nid) {
		return *k.self, nil
	}
	dist, err := node.CalDistance(nid, k.self.ID)
	if err != nil {
		golog.Error(err)
		return node.Node{}, nil
	}
	partion := dist.Partion()
	kq, ok := k.routes[partion]
	if ok {
		ok, n := kq.findOne(nid)
		if ok {
			return n, nil
		}
	}

	pslice := make([]int, len(k.routes))
	i := 0
	for key := range k.routes {
		pslice[i] = key
		i++
	}
	p := Partions{
		parts: pslice,
		base:  partion,
	}
	sort.Sort(p)
	for _, v := range p.parts {
		kq, ok := k.routes[v]
		if ok {
			ok, n := kq.findOne(nid)
			if ok {
				return n, nil
			}
		}
	}
	return node.Node{}, errors.New("NOT FOUND")
}
