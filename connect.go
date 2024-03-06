package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/xerrors"
	"log"
	"os"
	"strings"
)

// TODO more informative/debug-friendly logs: step-by-step, node ID

func processArgs(args []string) ([]peer.ID, [][]multiaddr.Multiaddr, error) {
	ids := make([]peer.ID, len(args))
	addrsByPeer := make([][]multiaddr.Multiaddr, len(args))
	for i, arg := range args {
		peerID, multiAddrs, err := processArg(arg)
		if err != nil {
			return nil, nil, xerrors.Errorf("could not process arg: %v", err)
		}
		ids[i] = peerID
		addrsByPeer[i] = multiAddrs
	}
	return ids, addrsByPeer, nil
}

func processArg(arg string) (peer.ID, []multiaddr.Multiaddr, error) {
	// parse peer ID and multi-addresses
	start, end := strings.Index(arg, "["), strings.Index(arg, "]")
	id, err := peer.Decode(arg[:start])
	if err != nil {
		return "", nil, xerrors.Errorf("could not decode Peer ID: %v", err)
	}
	addrs := strings.Split(arg[start+1:end], " ")
	multiAddrs := make([]multiaddr.Multiaddr, len(addrs))
	for i, addr := range addrs {
		mAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return "", nil, xerrors.Errorf("could not parse address %s: %v", addr, err)
		}
		multiAddrs[i] = mAddr
	}
	return id, multiAddrs, nil
}

func startHost(id peer.ID, addrs ...multiaddr.Multiaddr) (host.Host, error) {
	log.Printf("starting host %s at %v...\n", id, addrs)
	// load node's private key
	name := fmt.Sprintf("%v.txt", id)
	bytes, err := os.ReadFile(name)
	if err != nil {
		return nil, xerrors.Errorf("could not read %s: %v", name, err)
	}
	privKey, err := crypto.UnmarshalPrivateKey(bytes)
	if err != nil {
		return nil, xerrors.Errorf("could not deserialize the key: %v", err)
	}
	// create node
	h, err := libp2p.New(libp2p.ListenAddrs(addrs...), libp2p.Identity(privKey))
	if err != nil {
		return nil, xerrors.Errorf("could not create host %s: %v", id, err)
	}
	log.Printf("started host %s at %v\n", h.ID(), h.Addrs())
	return h, nil
}

func addPeers(h host.Host, peers []peer.ID, addrs [][]multiaddr.Multiaddr) error {
	log.Printf("adding peers to peer store...\n")
	for i, id := range peers {
		h.Peerstore().AddAddrs(id, addrs[i], peerstore.PermanentAddrTTL)
	}
	log.Printf("added peers\n")
	return nil
}

func connectPeers(h host.Host, p protocol.ID, handler network.StreamHandler) {
	// listen for incoming streams
	h.SetStreamHandler(p, handler)
	log.Printf("listening for streams...\n")
	// initiate outgoing streams
	log.Printf("opening streams...\n")
	for _, id := range h.Peerstore().Peers() {
		if id == h.ID() {
			continue
		}
		s, err := h.NewStream(context.Background(), id, p)
		if err != nil {
			// simply fail and expect the other node to connect to us later
			log.Printf("could not open stream to %s: %v\n", id, err)
			continue
		}
		log.Printf("opened stream to %s\n", id)
		go handler(s)
	}
}

func pingOnce(s network.Stream) {
	log.Printf("sending/receiving ping...\n")
	src, dst := s.Conn().LocalPeer(), s.Conn().RemotePeer()
	// send ping
	msg := createPing(src, dst, 1)
	_, err := s.Write([]byte(msg))
	if err != nil {
		log.Fatalf("could not write to stream: %v", err)
		return
	}
	fmt.Printf("sent ping: %s", msg)
	// listen for ping
	reader := bufio.NewReader(s)
	msg, err = reader.ReadString(pingDelimiter)
	if err != nil {
		log.Fatalf("could not read from stream: %v", err)
		return
	}
	fmt.Printf("received ping: %s", msg)
}
