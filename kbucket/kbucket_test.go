package kbucket

import (
	"kad/node"
	"testing"
)

func TestFind(t *testing.T) {
	id, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000000")
	n := node.Node{
		ID:    id,
		Addr:  "addr",
		State: node.NSNil,
	}
	k := New(&n)
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
		k.AddNode(v)
	}
	nid, _ := node.NewIDFromString("00000000-0000-0000-0000-000000000005")
	res, e := k.FindOne(nid)
	if e != nil {
		t.Error("[Kbucket.FindOne] ", e)
	}
	t.Log(res)
	if res.Addr != "addr3" {
		t.Error("[Kbucket.FindOne] closest one != 'addr3'")
	}
}
