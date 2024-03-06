package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"golang.org/x/xerrors"
	"log"
	"os"
)

// TODO keeping the same Peer ID across program executions requires persisting the secret key & using it to create host
// TODO %w failing xerrors.Errorf() inspection check
// TODO merge local.go & distributed.go to a single file

func createHost() (host.Host, error) {
	log.Printf("creating host...\n")

	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0/ws"))
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

func showAddrs(h host.Host) {
	fmt.Printf("Your address is: %v\n", h.Addrs())
	fmt.Printf("Run -connect \"%s%s\"\n", h.ID(), h.Addrs())
}
