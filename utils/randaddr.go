package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	// Currently only these ports are whitelisted on AWS security group!
	PortStart = 30001
	PortEnd   = 35000
	PortRange = PortEnd - PortStart
)

// Returns a randome port number in range [PortStart, PortEnd)
func RandPort() int {
	return PortStart + int(r.Intn(PortRange))
}

// Resolves the port number of a network address string if it ends with ":rand".
// If the string does not end with ":rand", the string is returned unchanged
func Resolve(s string) string {
	if strings.HasSuffix(s, ":rand") {
		s = strings.TrimSuffix(s, ":rand")
		s = fmt.Sprintf("%s:%d", s, RandPort())
	}
	return s
}

// A shortcut for Resolve("localhost:rand")
func Local() string {
	return fmt.Sprintf("localhost:%d", RandPort())
}
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
