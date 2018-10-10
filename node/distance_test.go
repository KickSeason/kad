package node

import "testing"

func TestPartion(t *testing.T) {
	b := []byte{0, 4}
	dis := newDistance(b)
	if dis.Partion() != 3 {
		t.Error(`distance{[]byte{0, 4}}.Partion() != 3`)
	}
}
