package kbucket

import (
	"fmt"
	"kad/node"

	"github.com/kataras/golog"
)

type KQue struct {
	que    []node.Node
	k      int
	bucket *Kbucket
}

func newKQue(k *Kbucket) KQue {
	return KQue{
		que:    make([]node.Node, 0, k.k),
		k:      k.k,
		bucket: k,
	}
}

func (kq *KQue) count() int {
	return len(kq.que)
}

func (kq *KQue) findN(nid node.NodeID, n int) ([]node.Node, error) {
	if kq.count() <= n {
		res := make([]node.Node, kq.count())
		for i, v := range kq.que {
			res[i] = v
		}
		return res, nil
	}
	li, err := findClosestN(n, nid, kq.que)
	fmt.Println(li)
	if err != nil {
		return []node.Node{}, err
	}
	return li, nil
}

func (kq *KQue) findOne(nid node.NodeID) (bool, node.Node) {
	if kq.count() == 0 {
		return false, node.Node{}
	}
	if kq.count() == 1 {
		return true, kq.que[0]
	}
	li, err := findClosestOne(nid, kq.que)
	if err != nil {
		golog.Error("[kq.findOne]", err)
		return false, node.Node{}
	}
	return true, li[0]
}

func (kq *KQue) has(n node.Node) bool {
	for _, v := range kq.que {
		if v.ID.Equal(n.ID) {
			return true
		}
	}
	return false
}
func (kq *KQue) update(n node.Node) {
	arr := []node.Node{}
	for _, v := range kq.que {
		if !v.ID.Equal(n.ID) {
			arr = append(arr, v)
		}
	}
	arr = append(arr, n)
	kq.que = arr
}
func (kq *KQue) updateAdd(n node.Node) {
	if kq.has(n) {
		kq.update(n)
		return
	}
	if kq.count() < kq.k {
		kq.que = append(kq.que, n)
		return
	}
	head := kq.que[0]
	if head.State == node.NSWaitPong {
		kq.update(head)
		return
	}
	if head.State == node.NSNil {
		head.State = node.NSWaitPong
		kq.que[0] = head
		kq.bucket.Ping <- head
	}
}

func (kq *KQue) remove(n node.Node) {
	if !kq.has(n) {
		return
	}
	arr := []node.Node{}
	for _, v := range kq.que {
		if !v.ID.Equal(n.ID) {
			arr = append(arr, v)
		}
	}
	kq.que = arr
}

func findClosestOne(nid node.NodeID, nodes []node.Node) ([]node.Node, error) {
	return findClosestN(1, nid, nodes)
}

func findClosestN(k int, nid node.NodeID, nodes []node.Node) ([]node.Node, error) {
	tmp := make([]node.Node, len(nodes))
	for i, v := range nodes {
		tmp[i] = v
	}
	pos, err := quickSort(nid, tmp, 0, len(nodes)-1)
	if err != nil {
		return []node.Node{}, err
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
			return []node.Node{}, err
		}
	}
	return tmp[:pos+1], nil
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
			d, err := node.CalDistance(nid, nodes[i].ID)
			if err != nil {
				return 0, err
			}
			if dis.Compare(d) <= 0 {
				break
			}
			i++
		}
		for i < j {
			d, err := node.CalDistance(nid, nodes[j].ID)
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
