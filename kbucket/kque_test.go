package kbucket

import (
	"kad/node"
	"testing"
)

func TestFindClosestN(t *testing.T) {
	id0, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000000")
	id1, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000001")
	id2, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000002")
	nodes := []node.Node{
		node.Node{
			ID:    id0,
			Addr:  "addr0",
			State: node.NSNil,
		},
		node.Node{
			ID:    id1,
			Addr:  "addr1",
			State: node.NSNil,
		},
		node.Node{
			ID:    id2,
			Addr:  "addr2",
			State: node.NSNil,
		},
	}
	nid, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000003")

	k := 1
	r, _ := findClosestN(k, nid, nodes)
	if len(r) != 1 {
		t.Error("[findClosestN] not enough k")
	}
	if r[0].Addr != "addr2" {
		t.Error("[findClosestN] most closest != '0002'")
	}

	k = 2
	r, _ = findClosestN(k, nid, nodes)

	if len(r) != 2 {
		t.Error("[findClosestN] not enough k")
	}
}
func TestUpdateAdd(t *testing.T) {
	id, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000000")
	n := node.Node{
		ID:    id,
		Addr:  "addr",
		State: node.NSNil,
	}
	k := New(&n)
	kq := newKQue(k)
	id0, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000001")
	id1, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000002")
	id2, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000003")
	id3, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000004")
	nodes := []node.Node{
		node.Node{
			ID:    id0,
			Addr:  "addr0",
			State: node.NSNil,
		},
		node.Node{
			ID:    id1,
			Addr:  "addr1",
			State: node.NSNil,
		},
		node.Node{
			ID:    id2,
			Addr:  "addr2",
			State: node.NSNil,
		},
		node.Node{
			ID:    id3,
			Addr:  "addr3",
			State: node.NSNil,
		},
	}
	for _, v := range nodes {
		kq.updateAdd(v)
	}
	t.Log(kq)
	if kq.count() != 4 {
		t.Error("[kQue.updateadd] count != 4")
	}
}
func TestFindOne(t *testing.T) {
	id, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000000")
	n := node.Node{
		ID:    id,
		Addr:  "addr",
		State: node.NSNil,
	}
	k := New(&n)
	kq := newKQue(k)
	id0, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000001")
	id1, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000002")
	id2, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000003")
	id3, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000004")
	nodes := []node.Node{
		node.Node{
			ID:    id0,
			Addr:  "addr0",
			State: node.NSNil,
		},
		node.Node{
			ID:    id1,
			Addr:  "addr1",
			State: node.NSNil,
		},
		node.Node{
			ID:    id2,
			Addr:  "addr2",
			State: node.NSNil,
		},
		node.Node{
			ID:    id3,
			Addr:  "addr3",
			State: node.NSNil,
		},
	}
	for _, v := range nodes {
		kq.updateAdd(v)
	}
	nid, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000005")
	res, e := kq.findN(nid, 3)
	if e != nil {
		t.Error("[kQue.findN] find error", e)
	}
	if len(res) != 3 {
		t.Error("[kQUe.findN] find count != 3")
	}
}
