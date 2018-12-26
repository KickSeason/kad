package kbs

import (
	"bytes"

	"github.com/KickSeason/kad/util"
)

//Distance distance between nodes
type Distance struct {
	distance []byte
}

func newDistance(d []byte) Distance {
	return Distance{
		distance: d,
	}
}

//CalDistance calculate distance between two nodes
func CalDistance(n NodeID, m NodeID) (d Distance, err error) {
	nbyte, e := n.ToByte()
	if e != nil {
		return d, e
	}
	mbyte, e := m.ToByte()
	if e != nil {
		return d, e
	}
	r, e := util.Xor(nbyte, mbyte)
	if e != nil {
		return d, e
	}
	return newDistance(r), nil
}

//Equal check wether two distances are equal
func (d Distance) Equal(dd Distance) bool {
	if len(d.distance) != len(dd.distance) {
		return false
	}
	return bytes.Equal(d.distance, dd.distance)
}

//Compare compare two distance
func (d Distance) Compare(dd Distance) int {
	if len(d.distance) != len(dd.distance) {
		return 0
	}
	return bytes.Compare(d.distance, dd.distance)
}

//Partion get the partion by distance
func (d Distance) Partion() int {
	p := len(d.distance)*8 - 1
	for _, v := range d.distance {
		if v == 0 {
			p -= 8
		} else {
			flag := 0x80
			for flag&int(v) == 0 {
				flag >>= 1
				p--
			}
		}
	}
	p++
	return p
}
