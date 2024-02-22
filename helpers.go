package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	ma "github.com/multiformats/go-multiaddr"
)

type Host struct {
	host.Host
}

func (h Host) String() string {
	return getMultiAddress(h)
}

func getMultiAddress(h host.Host) string {
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", h.ID()))
	addr := h.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}
