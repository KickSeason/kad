package kbs

import (
	"testing"
)

func TestFind(t *testing.T) {
	id, _ := NewIDFromString("00000000-0000-0000-0000-000000000000")
	n := Node{
		ID:    id,
		Addr:  "addr",
		State: NSNil,
	}
	k := New(&n)
	id0, _ := NewIDFromString("00000000-0000-0000-0000-000000000001")
	id1, _ := NewIDFromString("00000000-0000-0000-0000-000000000002")
	id2, _ := NewIDFromString("00000000-0000-0000-0000-000000000003")
	id3, _ := NewIDFromString("00000000-0000-0000-0000-000000000004")
	nodes := []Node{
		Node{
			ID:    id0,
			Addr:  "addr0",
			State: NSNil,
		},
		Node{
			ID:    id1,
			Addr:  "addr1",
			State: NSNil,
		},
		Node{
			ID:    id2,
			Addr:  "addr2",
			State: NSNil,
		},
		Node{
			ID:    id3,
			Addr:  "addr3",
			State: NSNil,
		},
	}
	for _, v := range nodes {
		k.AddNode(v)
	}
	nid, _ := NewIDFromString("00000000-0000-0000-0000-000000000005")
	res, e := k.FindOne(nid)
	if e != nil {
		t.Error("[FindOne] ", e)
	}
	t.Log(res)
	if res.Addr != "addr3" {
		t.Error("[FindOne] closest one != 'addr3'")
	}
}
