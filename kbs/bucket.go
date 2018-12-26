package kbs

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kataras/golog"
)

//bucket a queue to store nodes
type bucket struct {
	que   []Node
	k     int
	kbs   *KBS
	mutex sync.RWMutex
}

func newbucket(k *KBS) bucket {
	return bucket{
		que: make([]Node, 0, k.k),
		k:   k.k,
		kbs: k,
	}
}

func (b *bucket) count() int {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return len(b.que)
}

func (b *bucket) findN(nid NodeID, n int) ([]Node, error) {
	if b.count() <= n {
		res := make([]Node, b.count())
		b.mutex.Lock()
		for i, v := range b.que {
			res[i] = v
		}
		b.mutex.Unlock()
		return res, nil
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	li, err := findClosestN(n, nid, b.que)
	fmt.Println(li)
	if err != nil {
		return []Node{}, err
	}
	return li, nil
}

func (b *bucket) findOne(nid NodeID) (bool, Node) {
	if b.count() == 0 {
		return false, Node{}
	}
	if b.count() == 1 {
		return true, b.que[0]
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	li, err := findClosestOne(nid, b.que)
	if err != nil {
		golog.Error("[b.findOne]", err)
		return false, Node{}
	}
	return true, li[0]
}

func (b *bucket) has(n Node) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, v := range b.que {
		if v.ID.Equal(n.ID) {
			return true
		}
	}
	return false
}
func (b *bucket) update(n Node) {
	arr := []Node{}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, v := range b.que {
		if !v.ID.Equal(n.ID) {
			arr = append(arr, v)
		}
	}
	if n.state == nsping {
		n.state = nsnil
	}
	arr = append(arr, n)
	b.que = arr
}
func (b *bucket) replace(nid NodeID, n Node) {
	arr := []Node{}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, v := range b.que {
		if !v.ID.Equal(nid) {
			arr = append(arr, v)
		}
	}
	arr = append(arr, n)
	b.que = arr
}
func (b *bucket) add(n Node) {
	if b.has(n) {
		b.update(n)
		return
	}
	if b.count() < b.k {
		b.mutex.Lock()
		b.que = append(b.que, n)
		b.mutex.Unlock()
		b.kbs.send(MailFindAsync, []interface{}{b.kbs.Self.ID, n})
		return
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()
	head := b.que[0]
	if s := head.state & nsping; s != 0 {
		return
	}
	head.state |= nsping
	b.que[0] = head
	_, err := b.kbs.send(MailPingSync, []interface{}{head})
	if err != nil {
		head.state &= ^nsping
		b.update(head)
		return
	}
	b.replace(head.ID, n)
}

func (b *bucket) remove(n Node) {
	if !b.has(n) {
		return
	}
	arr := []Node{}
	for _, v := range b.que {
		if !v.ID.Equal(n.ID) {
			arr = append(arr, v)
		}
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.que = arr
}

func (b *bucket) tojson() string {
	bytes, err := json.Marshal(b.que)
	if err != nil {
		golog.Error("[bucket.tojson] ", err)
		return ""
	}
	return string(bytes)
}

func findClosestOne(nid NodeID, nodes []Node) ([]Node, error) {
	return findClosestN(1, nid, nodes)
}

func findClosestN(k int, nid NodeID, nodes []Node) ([]Node, error) {
	tmp := make([]Node, len(nodes))
	for i, v := range nodes {
		tmp[i] = v
	}
	pos, err := quickSort(nid, tmp, 0, len(nodes)-1)
	if err != nil {
		return []Node{}, err
	}
	start := 0
	end := len(tmp) - 1
	for pos != k-1 {
		if pos < k-1 {
			start = pos + 1
		} else {
			end = pos - 1
		}
		pos, err = quickSort(nid, tmp, start, end)
		if err != nil {
			return []Node{}, err
		}
	}
	return tmp[:pos+1], nil
}

func quickSort(nid NodeID, nodes []Node, start int, end int) (int, error) {
	flag := nodes[start]
	dis, err := CalDistance(flag.ID, nid)
	if err != nil {
		golog.Error(err)
		return 0, err
	}
	i := start
	j := end
	for i < j {
		for i < j {
			d, err := CalDistance(nid, nodes[i].ID)
			if err != nil {
				return 0, err
			}
			if dis.Compare(d) <= 0 {
				break
			}
			i++
		}
		for i < j {
			d, err := CalDistance(nid, nodes[j].ID)
			if err != nil {
				return 0, err
			}
			if 0 <= dis.Compare(d) {
				break
			}
			j--
		}
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
	return i, nil
}
