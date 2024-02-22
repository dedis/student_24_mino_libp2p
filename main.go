package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// TODO: hostsStr(), peerStoreStr()
// TODO: host package doc
// TODO: extract establishing connections to separate method
// TODO: debug - displaying received pings stops after 2 messages
// TODO: context.WithCancel() purpose, needed?
// TODO: change fmt to log
// TODO example echo host: tests

const (
	ProtocolPing = protocol.ID("my/ping/1.0.0")
	N            = 3
)

// set up N nodes to send random pings to each other & display received pings
func main() {
	nodes, err := createNodes(N)
	if err != nil {
		panic(err)
	}

	startPingListeners(nodes)

	fillPeerStores(nodes)

	streams, err := openPingStreams(nodes)
	if err != nil {
		panic(err)
	}

	fmt.Printf("open streams: %v\n", streams)

	err = sendRandomPings(nodes, streams)
	if err != nil {
		panic(err)
	}
}

// create N nodes
func createNodes(n int) ([]Host, error) {
	fmt.Printf("creating nodes...\n")

	nodes := make([]Host, 0, n)
	for i := 0; i < n; i++ {
		node, err := libp2p.New() // by default listen on localhost random TCP port
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes, Host{node})
	}

	fmt.Printf("created nodes: %v\n", nodes)
	return nodes, nil
}

// listen for pings on all nodes & display received pings
func startPingListeners(nodes []Host) {
	// on all nodes
	for ni, node := range nodes {
		// listen for pings
		node.SetStreamHandler(ProtocolPing, func(stream network.Stream) {
			// read pings
			line, err := bufio.NewReader(stream).ReadString('\n')
			if err != nil {
				panic(err)
			}
			// display pings
			fmt.Printf("received: %s", line)
		})
		fmt.Printf("listening for pings on %d/%d nodes...\n", ni+1, len(nodes))
	}
}

// add the address of a peer to my peer store
func fillPeerStores(nodes []Host) {
	fmt.Printf("filling peer stores...\n")

	//	each node needs to know all other nodes
	for i, node := range nodes {
		for _, peer := range nodes {
			if node.ID() != peer.ID() {
				node.Peerstore().AddAddrs(peer.ID(), peer.Addrs(), peerstore.PermanentAddrTTL)
			}
		}
		fmt.Printf("filled %d/%d peer stores: {node: %v, peer store: size %d}\n", i+1, len(nodes), node.ID(), len(node.Peerstore().Peers()))
	}
}

// open streams for ping protocol (and create connections if haven't already) between all pairs of nodes
func openPingStreams(nodes []Host) ([][]network.Stream, error) {
	fmt.Printf("opening ping streams...\n")

	streams := make([][]network.Stream, len(nodes))
	for i := range streams {
		streams[i] = make([]network.Stream, len(nodes))
	}

	nBiStreams := len(nodes) * (len(nodes) - 1) / 2
	count := 0
	// between every pair of nodes
	for si, src := range nodes {
		for di, dst := range nodes {
			// don't open stream to node itself or if already opened
			if si == di || streams[si][di] != nil {
				continue
			}
			// open stream (note: a connection is bidirectional and is created if not already)
			stream, err := src.NewStream(context.Background(), dst.ID(), ProtocolPing)
			if err != nil {
				return streams, err
			}
			// store the bidirectional stream
			streams[si][di], streams[di][si] = stream, stream
			count++
			fmt.Printf("opened %d/%d ping streams\n", count, nBiStreams)
		}
	}

	return streams, nil
}

// send ping from one random node to another indefinitely
func sendRandomPings(nodes []Host, streams [][]network.Stream) error {
	fmt.Printf("sending random pings...\n")

	for count := 1; ; {
		// between random pair of nodes
		si := rand.Intn(len(nodes))
		di := rand.Intn(len(nodes))
		if si != di {
			// send ping
			msg := fmt.Sprintf("ping #%d from %v to %v\n", count, nodes[si].ID(), nodes[di].ID())
			_, err := streams[si][di].Write([]byte(msg))
			if err != nil {
				return err
			}
			fmt.Printf("sent random ping #%d\n", count)
			count++
			// wait for a fixed interval
			time.Sleep(3 * time.Second)
		}
	}
}
