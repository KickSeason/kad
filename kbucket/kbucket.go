package kbucket

import (
	"errors"
	"sort"
	"time"

	"github.com/kataras/golog"
)

type (
	NoteType uint8
	Note     struct {
		Type   NoteType
		Arg    interface{}
		result chan interface{}
	}
	Kbucket struct {
		routes   map[int]KQue
		Self     *Node
		store    *Storage
		k        int
		alpha    int
		ticker   *time.Ticker
		receiver chan Note
		Sender   chan Note
	}
)

const (
	nDelNode NoteType = 0x01
	nAddNode NoteType = 0x02
	nFindOne NoteType = 0x03
	nFind    NoteType = 0x04
	nStore   NoteType = 0x05
	NPing
)
const (
	kcount = 8
	alpha  = 3
	ticktm = 5 * time.Second
)

//New create a kbucket
func New(local *Node) *Kbucket {
	k := &Kbucket{
		routes:   make(map[int]KQue, 64),
		Self:     local,
		store:    NewStorage(),
		k:        kcount,
		alpha:    alpha,
		receiver: make(chan Note),
		Sender:   make(chan Note),
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
		case msg := <-k.receiver:
			switch msg.Type {
			case nAddNode:
				n := msg.Arg.(Node)
				k.add(n)
			case nDelNode:
				n := msg.Arg.(Node)
				k.remove(n)
			case nFind:
				nid := msg.Arg.(NodeID)
				k.find(nid, msg.result)
			case nFindOne:
				nid := msg.Arg.(NodeID)
				k.findOne(nid, msg.result)
			case nStore:
				kv := msg.Arg.(struct {
					key   string
					value string
				})
				k.storeKV(kv.key, kv.value)
			}
		}
	}
}

//AddNode to add a node
func (k *Kbucket) AddNode(n Node) {
	k.receiver <- Note{
		Type: nAddNode,
		Arg:  n,
	}
}
func (k *Kbucket) add(n Node) {
	distance, err := CalDistance(n.ID, k.Self.ID)
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
func (k *Kbucket) RemoveNode(n Node) {
	k.receiver <- Note{
		Type: nDelNode,
		Arg:  n,
	}
}
func (k *Kbucket) remove(n Node) {
	distance, err := CalDistance(n.ID, k.Self.ID)
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
func (k *Kbucket) Find(nid NodeID) (ns []Node, err error) {
	phone := make(chan interface{})
	k.receiver <- Note{
		Type:   nFind,
		Arg:    nid,
		result: phone,
	}
	result, ok := <-phone
	if !ok {
		return ns, errors.New("Failed")
	}
	return result.([]Node), nil
}

//Find find alpha nodes that are closest to the nid
func (k *Kbucket) find(nid NodeID, phone chan interface{}) {
	defer close(phone)
	var ns []Node
	if k.Self.ID.Equal(nid) {
		return
	}
	dist, err := CalDistance(nid, k.Self.ID)
	if err != nil {
		golog.Error(err)
		return
	}
	partion := dist.Partion()
	if kq, ok := k.routes[partion]; ok {
		ns, err = kq.findN(nid, k.alpha)
		if err != nil {
			return
		}
	}
	if k.alpha <= len(ns) {
		return
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
				return
			}
			for _, v := range res {
				ns = append(ns, v)
			}
			if k.alpha <= len(ns) {
				break
			}
		}
	}
	phone <- ns
}

func (k *Kbucket) FindOne(nid NodeID) (Node, error) {
	phone := make(chan interface{})
	k.receiver <- Note{
		Type:   nFindOne,
		Arg:    nid,
		result: phone,
	}
	result, ok := <-phone
	if !ok {
		return Node{}, errors.New("Failed")
	}
	return result.(Node), nil
}

func (k *Kbucket) findOne(nid NodeID, phone chan interface{}) (Node, error) {
	if k.Self.ID.Equal(nid) {
		return *k.Self, nil
	}
	dist, err := CalDistance(nid, k.Self.ID)
	if err != nil {
		golog.Error(err)
		return Node{}, nil
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
	return Node{}, errors.New("NOT FOUND")
}

func (k *Kbucket) Store(key, value string) {
	k.receiver <- Note{
		Type: nStore,
		Arg: struct {
			key   string
			value string
		}{key, value},
	}
}
func (k *Kbucket) storeKV(key, value string) {
	k.store.Put(key, value)
}

func (k *Kbucket) send(nt NoteType, data interface{}) {
	switch nt {
	case NPing:
		k.Sender <- Note{
			Type: nt,
			Arg:  data.(Node),
		}
	}
}
