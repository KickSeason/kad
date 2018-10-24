package util

import (
	"github.com/kataras/golog"
)

type AbDistance interface {
}
type AbNode interface {
}

func SortAsDistance(n AbNode, ns []AbNode,
	CalDist func(a, b AbNode) (AbDistance, error),
	Compare func(a, b AbDistance) int) ([]AbNode, error) {
	return FindClosestN(len(ns), n, ns, CalDist, Compare)
}

func FindClosestN(k int, n AbNode, ns []AbNode,
	CalDist func(a, b AbNode) (AbDistance, error),
	Compare func(a, b AbDistance) int) ([]AbNode, error) {
	tmp := make([]AbNode, len(ns))
	for i, v := range ns {
		tmp[i] = v
	}
	pos, err := quickSort(n, tmp, 0, len(ns)-1, CalDist, Compare)
	if err != nil {
		return []AbNode{}, err
	}
	start := 0
	end := len(tmp) - 1
	for pos != k {
		if pos < k {
			start = pos + 1
		} else {
			end = pos - 1
		}
		pos, err = quickSort(n, tmp, start, end, CalDist, Compare)
		if err != nil {
			return []AbNode{}, err
		}
	}
	return tmp[:pos], nil
}

func quickSort(n AbNode, ns []AbNode, start int, end int,
	CalDist func(a, b AbNode) (AbDistance, error),
	Compare func(a, b AbDistance) int) (int, error) {
	flag := ns[start]
	dis, err := CalDist(n, flag)
	if err != nil {
		golog.Error(err)
		return 0, err
	}
	i := start
	j := end
	for i < j {
		for i < j {
			d, err := CalDist(n, ns[i])
			if err != nil {
				return 0, err
			}
			if Compare(dis, d) <= 0 {
				break
			}
			i++
		}
		for i < j {
			d, err := CalDist(n, ns[i])
			if err != nil {
				return 0, err
			}
			if 0 <= Compare(dis, d) {
				break
			}
			j--
		}
		ns[i], ns[j] = ns[j], ns[i]
	}
	return i, nil
}
