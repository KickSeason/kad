package util

import (
	"fmt"
	"math"
	"testing"
)

type node int
type dist int

func caldistance(n, m node) (dist, error) {
	return dist(math.Abs(float64(n - m))), nil
}
func compare(d, e dist) int {
	if e < d {
		return 1
	} else if d < e {
		return -1
	}
	return 0
}
func TestSort(t *testing.T) {
	n := node(0)
	ns := []node{1, 5, 3, 2, 4}
	r, _ := FindClosestN(3, n, ns, caldistance, compare)
	fmt.Println(r)
}
