package kbs

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/kataras/golog"
)

type (
	notetype uint8
	MailType uint8
	note     struct {
		typ    notetype
		arg    interface{}
		result chan interface{}
	}
	Mail struct {
		Type   MailType
		Arg    []interface{}
		Result chan interface{}
	}
	KbConfig struct {
		Seeds   []string
		LocalIP net.IP
		Port    uint32
		ID      NodeID
	}
	KBS struct {
		config   *KbConfig
		routes   map[int]bucket
		Self     *Node
		store    *Storage
		k        int
		alpha    int
		ticker   *time.Ticker
		receiver chan note
		Outbox   chan Mail
	}
)

const (
	nDelNode notetype = 0x01
	nAddNode notetype = 0x02
	nFindOne notetype = 0x03
	nFind    notetype = 0x04
	nStore   notetype = 0x05
	nGet     notetype = 0x06
)

const (
	MailPingSync  MailType = 0x16
	MailFindSync  MailType = 0x17
	MailPingAsync MailType = 0x96
	MailFindAsync MailType = 0x97
)
const (
	kcount = 8
	alpha  = 3
	ticktm = 5 * time.Second
)

//New create a KBS
func NewKBS(config *KbConfig) *KBS {
	n := NewNode(config.ID, config.LocalIP, config.Port)
	k := &KBS{
		config:   config,
		routes:   make(map[int]bucket, 64),
		Self:     &n,
		store:    NewStorage(),
		k:        kcount,
		alpha:    alpha,
		receiver: make(chan note),
		Outbox:   make(chan Mail),
	}
	k.ticker = time.NewTicker(ticktm)
	return k
}

func (k *KBS) Start() {
	go k.run()
}

func (k *KBS) run() {
	for {
		select {
		case <-k.ticker.C:
			//golog.Info("[KBS.run] routes: ", k.ToJson())
		case msg := <-k.receiver:
			switch msg.typ {
			case nAddNode:
				n := msg.arg.(Node)
				k.add(n)
			case nDelNode:
				n := msg.arg.(Node)
				k.remove(n)
			case nFind:
				nid := msg.arg.(NodeID)
				k.find(nid, msg.result)
			case nFindOne:
				nid := msg.arg.(NodeID)
				k.findOne(nid, msg.result)
			case nStore:
				kv := msg.arg.(struct {
					key   string
					value string
				})
				k.storeKV(kv.key, kv.value)
			case nGet:
				key := msg.arg.(string)
				k.get(key, msg.result)
			}
		}
	}
}

//AddNode to add a node
func (k *KBS) AddNode(n Node) {
	k.receiver <- note{
		typ: nAddNode,
		arg: n,
	}
}
func (k *KBS) add(n Node) {
	if n.ID.Equal(k.Self.ID) {
		return
	}
	distance, err := CalDistance(n.ID, k.Self.ID)
	if err != nil {
		golog.Error(err)
	}
	partion := distance.Partion()
	var bk bucket
	if _, ok := k.routes[partion]; !ok {
		bk = newbucket(k)
	} else {
		bk = k.routes[partion]
	}
	qptr := &bk
	qptr.add(n)
	k.routes[partion] = bk
	return
}

//RemoveNode to remove a node
func (k *KBS) RemoveNode(n Node) {
	k.receiver <- note{
		typ: nDelNode,
		arg: n,
	}
}
func (k *KBS) remove(n Node) {
	distance, err := CalDistance(n.ID, k.Self.ID)
	if err != nil {
		golog.Error(err)
	}
	partion := distance.Partion()
	if _, ok := k.routes[partion]; !ok {
		return
	}
	bk := k.routes[partion]
	bk.remove(n)
	k.routes[partion] = bk
	return
}
func (k *KBS) Find(nid NodeID) (ns []Node, err error) {
	phone := make(chan interface{})
	k.receiver <- note{
		typ:    nFind,
		arg:    nid,
		result: phone,
	}
	result, ok := <-phone
	if !ok {
		return ns, errors.New("Failed")
	}
	return result.([]Node), nil
}

//Find find alpha nodes that are closest to the nid
func (k *KBS) find(nid NodeID, phone chan interface{}) {
	defer close(phone)
	var ns []Node
	if k.Self.ID.Equal(nid) {
		phone <- []interface{}{nid}
		return
	}
	dist, err := CalDistance(nid, k.Self.ID)
	if err != nil {
		golog.Error(err)
		return
	}
	partion := dist.Partion()
	if bk, ok := k.routes[partion]; ok {
		ns, err = bk.findN(nid, k.alpha)
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
		bk, ok := k.routes[v]
		if ok {
			res, err := bk.findN(nid, k.alpha-len(ns))
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

func (k *KBS) FindOne(nid NodeID) (Node, error) {
	phone := make(chan interface{})
	k.receiver <- note{
		typ:    nFindOne,
		arg:    nid,
		result: phone,
	}
	result, ok := <-phone
	if !ok {
		return Node{}, errors.New("Failed")
	}
	return result.(Node), nil
}

func (k *KBS) findOne(nid NodeID, phone chan interface{}) (Node, error) {
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

func (k *KBS) KeyToID(key string) NodeID {
	return NodeID{}
}

func (k *KBS) Store(key, value string) {
	k.receiver <- note{
		typ: nStore,
		arg: struct {
			key   string
			value string
		}{key, value},
	}
}

func (k *KBS) storeKV(key, value string) {
	k.store.Put(key, value)
}

func (k *KBS) Get(key string) (string, error) {
	phone := make(chan interface{})
	k.receiver <- note{
		typ:    nGet,
		arg:    key,
		result: phone,
	}
	result, ok := <-phone
	if !ok {
		return "", errors.New("Not Found")
	}
	return result.(string), nil
}

func (k *KBS) get(key string, result chan interface{}) {
	defer close(result)
	value, err := k.store.Get(key)
	if err != nil {
		result <- value
		return
	}
}

func (k *KBS) send(mt MailType, data []interface{}) (interface{}, error) {
	switch mt {
	case MailPingSync:
		mail := Mail{
			Type:   mt,
			Arg:    data,
			Result: make(chan interface{}),
		}
		k.Outbox <- mail
		nid, ok := <-mail.Result
		if !ok {
			return nil, errors.New("Ping Failed")
		}
		return nid, nil
	case MailPingAsync:
		mail := Mail{
			Type:   mt,
			Arg:    data,
			Result: nil,
		}
		k.Outbox <- mail
	case MailFindSync:
		mail := Mail{
			Type:   mt,
			Arg:    data,
			Result: make(chan interface{}),
		}
		k.Outbox <- mail
		ns, ok := <-mail.Result
		if !ok {
			return nil, errors.New("Find Failed")
		}
		return ns, nil
	case MailFindAsync:
		mail := Mail{
			Type:   mt,
			Arg:    data,
			Result: nil,
		}
		k.Outbox <- mail
	}
	return nil, nil
}

func (k *KBS) ToJson() string {
	jstr := "{["
	for k, v := range k.routes {
		jstr += fmt.Sprintf(`{"distance": %d, "nodes": %s}`, k, v.tojson())
	}
	jstr += "]}"
	return string(jstr)
}
