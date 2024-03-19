package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	ws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/xerrors"
	"log"
	"os"
)

// TODO keeping the same Peer ID across program executions requires persisting the secret key & using it to create host
// TODO %w failing xerrors.Errorf() inspection check
// TODO merge local.go & distributed.go to a single file

func createHost() (host.Host, error) {
	log.Printf("creating host...\n")

	// TODO: - If no security transport is provided, the host uses the go-libp2p's noise and/or tls encrypted transport to encrypt all traffic;
	// TODO: To stop/shutdown the returned libp2p node, the user needs to cancel the passed context and call `Close` on the returned Host.
	// listen on any interface and a random port for websocket connections
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0/ws"), libp2p.Transport(ws.New))
	if err != nil {
		return nil, xerrors.Errorf("could not create host: %v", err)
	}

	log.Printf("created host %v\n", h.ID())
	return h, nil
}

func storeIdentity(h host.Host) error {
	log.Printf("storing identity to file...\n")

	name := fmt.Sprintf("%s.txt", h.ID())
	f, err := os.Create(name)
	if err != nil {
		return xerrors.Errorf("could not create file: %v", err)
	}
	defer f.Close()

	privKey := h.Peerstore().PrivKey(h.ID())
	bytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return xerrors.Errorf("could not serialize the key: %v", err)
	}
	_, err = f.Write(bytes)
	if err != nil {
		return xerrors.Errorf("could not write key to file: %v", err)
	}

	log.Printf("stored identity to file %v\n", f.Name())
	return nil
}

// TODO update README
func showAddrs(h host.Host) {
	listening := fmt.Sprintf("\"%s%s\"", h.ID(), h.Addrs())
	nonLocal := fmt.Sprintf("\"%s%s\"", h.ID(), getNonLocalAddrs(h))
	fmt.Printf("Your listening addresses: %v\n", listening)
	fmt.Printf("Your non-local addresses: %v\n", nonLocal)
	fmt.Printf("Run -connect \"listeningAddrs\" \"nonLocalAddrs(dest1)\" \"nonLocalAddrs(dest2)\" ...\n")
}

// TODO right way to do this with net/libp2p package?
func getNonLocalAddrs(h host.Host) []multiaddr.Multiaddr {
	var nonLocal []multiaddr.Multiaddr
	for _, addr := range h.Addrs() {
		ip, err := addr.ValueForProtocol(multiaddr.P_IP4)
		if err != nil {
			log.Printf("could not get ip4 addr: %s", err)
			continue
		}
		if ip != "127.0.0.1" {
			nonLocal = append(nonLocal, addr)
		}
	}
	return nonLocal
}
