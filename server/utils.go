package server

import "math/rand"

// pick a random server
func getLatestServerIndex(servers []*ServerStatus) int {
	var maxRev uint64 = 0
	s := make([]int, 0)

	for _, v := range servers {
		if v.Revision >= maxRev {
			maxRev = v.Revision
		}
	}
	for i, v := range servers {
		if v.Revision >= maxRev {
			s = append(s, i)
		}
	}

	return s[rand.Intn(len(s))]
}
