package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
)

// MustParsePeerList takes a newline-delimited file and parses it into a []Peer. Panics.
func MustParsePeerList(filename string) []Peer {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	raw := string(bytes)

	lines := strings.Split(raw, "\n")

	peers := make([]Peer, 0, len(lines))
	for _, line := range lines {
		u, err := url.Parse(line)
		fmt.Println(line)
		if err != nil {
			panic(err)
		}
		peers = append(peers, Peer{URL: u})
	}

	return peers
}
