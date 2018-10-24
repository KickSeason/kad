package kbucket

import "math"

type Partions struct {
	parts []int
	base  int
}

func (p Partions) Len() int {
	return len(p.parts)
}

func (p Partions) Less(i, j int) bool {
	d1 := math.Abs(float64(p.parts[i] - p.base))
	d2 := math.Abs(float64(p.parts[j] - p.base))
	return d1 < d2
}

func (p Partions) Swap(i, j int) {
	p.parts[i], p.parts[j] = p.parts[j], p.parts[i]
}
