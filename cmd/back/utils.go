package main

func GetrandomAddresses(n int) []string {
	backs := make(map[string]bool)
	i := 0
	for i < n {
		addr := Local()
		if _, ok := backs[addr]; !ok {
			backs[addr] = true
			i++
		}
	}

	backends := make([]string, 0)
	for k, _ := range backs {
		backends = append(backends, k)
	}
	return backends
}
