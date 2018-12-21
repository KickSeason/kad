package kbucket

import (
	"testing"
)

func TestPartion(t *testing.T) {
	b := []byte{0, 4}
	dis := newDistance(b)
	if dis.Partion() != 3 {
		t.Error(`distance{[]byte{0, 4}}.Partion() != 3`)
	}
}

func TestCompare(t *testing.T) {
	id0, _ := NewIDFromString("00000000-0000-0000-0000-000000000000")
	id1, _ := NewIDFromString("00000000-0000-0000-0000-000000000001")
	id2, _ := NewIDFromString("00000000-0000-0000-0000-000000000002")
	d1, _ := CalDistance(id2, id0)
	d2, _ := CalDistance(id2, id1)

	if 0 <= d1.Compare(d2) {
		t.Error("[distance.compare]distance compare failed.")
	}
}
