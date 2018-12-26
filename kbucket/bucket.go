package kbs

import (
	"fmt"

	"github.com/kataras/golog"
)

//bucket a queue to store nodes
type bucket struct {
	que    []Node
	k      int
	bucket *KBS
}

func newbucket(k *KBS) bucket {
	return bucket{
		que:    make([]Node, 0, k.k),
		k:      k.k,
		bucket: k,
	}
}

func (kq *bucket) count() int {
	return len(kq.que)
}

func (kq *bucket) findN(nid NodeID, n int) ([]Node, error) {
	if kq.count() <= n {
		res := make([]Node, kq.count())
		for i, v := range kq.que {
			res[i] = v
		}
		return res, nil
	}
	li, err := findClosestN(n, nid, kq.que)
	fmt.Println(li)
	if err != nil {
		return []Node{}, err
	}
	return li, nil
}

func (kq *bucket) findOne(nid NodeID) (bool, Node) {
	if kq.count() == 0 {
		return false, Node{}
	}
	if kq.count() == 1 {
		return true, kq.que[0]
	}
	li, err := findClosestOne(nid, kq.que)
	if err != nil {
		golog.Error("[kq.findOne]", err)
		return false, Node{}
	}
	return true, li[0]
}

func (kq *bucket) has(n Node) bool {
	for _, v := range kq.que {
		if v.ID.Equal(n.ID) {
			return true
		}
	}
	return false
}
func (kq *bucket) update(n Node) {
	arr := []Node{}
	for _, v := range kq.que {
		if !v.ID.Equal(n.ID) {
			arr = append(arr, v)
		}
	}
	if n.state == nsping {
		n.state = nsnil
	}
	arr = append(arr, n)
	kq.que = arr
}
func (kq *bucket) add(n Node) {
	if kq.has(n) {
		kq.update(n)
		return
	}
	if kq.count() < kq.k {
		kq.que = append(kq.que, n)
		return
	}
	head := kq.que[0]
	if head.state == nsping {
		return
	}
	if head.state == nsnil {
		head.state = nsping
		kq.que[0] = head
		kq.bucket.send(MailPing, []interface{}{head})
	}
}

func (kq *bucket) remove(n Node) {
	if !kq.has(n) {
		return
	}
	arr := []Node{}
	for _, v := range kq.que {
		if !v.ID.Equal(n.ID) {
			arr = append(arr, v)
		}
	}
	kq.que = arr
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
