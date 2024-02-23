package main

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

func getMultiAddress(h host.Host) string {
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", h.ID()))
	addr := h.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

func getPeers(ps peerstore.Peerstore) string {
	peers := make([]string, len(ps.Peers()))
	for i, id := range ps.Peers() {
		addresses := ps.Addrs(id)
		peers[i] = fmt.Sprintf("{peer: %s, addresses: %v}", id, addresses)
	}
	return fmt.Sprintf("[%s]", strings.Join(peers, ", "))
}
