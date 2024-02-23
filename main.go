package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// TODO: review, refactor, write docs
// TODO: extract streamHandler to displayPings() (represents listener vs dialer sendRandomPings())
// TODO: debug - displaying received pings stops after 2 messages
// TODO: stream doc: multiplexed = separate R vs W channels underneath?
// TODO: #230 Create a buffered stream so that read and writes are non-blocking??
// TODO: host package doc
// TODO: context.WithCancel() purpose, needed?
// TODO: change fmt to log
// TODO example echo host: tests

const (
	ProtocolPing = protocol.ID("my/ping/1.0.0")
	N            = 2
)

// set up N nodes to send random pings to each other & display received pings
func main() {
	nodes, err := createNodes(N)
	if err != nil {
		panic(err)
	}
	// as listeners: receive pings
	startPingListeners(nodes)
	// as dialers: send pings
	fillPeerStores(nodes)
	writers, err := openPingStreams(nodes)
	if err != nil {
		panic(err)
	}
	err = sendRandomPings(nodes, writers)
	if err != nil {
		panic(err)
	}
}

// create N nodes
func createNodes(n int) ([]host.Host, error) {
	fmt.Printf("creating nodes...\n")

	nodes := make([]host.Host, 0, n)
	for i := 0; i < n; i++ {
		node, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes, node)
		fmt.Printf("created %d/%d nodes: %v\n", i+1, n, getMultiAddress(node))
	}

	return nodes, nil
}

// listen for pings indefinitely on all nodes & display received pings
func startPingListeners(nodes []host.Host) {
	for i, node := range nodes {
		node.SetStreamHandler(ProtocolPing, func(s network.Stream) {
			reader := bufio.NewReader(s)
			for {
				msg, err := reader.ReadString('\n')
				if err != nil {
					panic(err)
				}
				fmt.Printf("received: %s", msg)
			}
		})
		fmt.Printf("listening for pings on %d/%d nodes...\n", i+1, len(nodes))
	}
}

// add the addresses of all peers to each node's peer store
func fillPeerStores(nodes []host.Host) {
	fmt.Printf("filling peer stores...\n")

	//	each node needs to know all other nodes
	for i, node := range nodes {
		for _, peer := range nodes {
			if node.ID() == peer.ID() {
				continue
			}
			node.Peerstore().AddAddrs(peer.ID(), peer.Addrs(), peerstore.PermanentAddrTTL)
		}
		fmt.Printf("filled %d/%d peer stores: {node: %v, peer store: %s}\n", i+1, len(nodes), node.ID(), getPeers(node.Peerstore()))
	}
}

// open streams for ping protocol (and connections if not already) between all pairs of nodes
func openPingStreams(nodes []host.Host) ([][]*bufio.Writer, error) {
	fmt.Printf("opening ping streams...\n")

	writers := make([][]*bufio.Writer, len(nodes))
	totalStreams := len(nodes) * (len(nodes) - 1)
	count := 0
	// between every pair of nodes
	for si, src := range nodes {
		writers[si] = make([]*bufio.Writer, len(nodes))
		for di, dst := range nodes {
			// don't open stream to node itself
			if si == di {
				continue
			}
			// open stream (note: a connection is bidirectional and is created if not already from either side)
			s, err := src.NewStream(context.Background(), dst.ID(), ProtocolPing)
			if err != nil {
				return writers, err
			}
			// store buffered writer over stream for non-blocking writes
			writers[si][di] = bufio.NewWriter(s)
			count++
			fmt.Printf("opened %d/%d ping streams\n", count, totalStreams)
		}
	}

	return writers, nil
}

// send ping from one random node to another indefinitely
func sendRandomPings(nodes []host.Host, writers [][]*bufio.Writer) error {
	fmt.Printf("sending random pings...\n")

	for count := 1; ; {
		// between a random pair of nodes
		si := rand.Intn(len(nodes))
		di := rand.Intn(len(nodes))
		if si == di {
			continue
		}
		// send ping
		src, dst, writer := nodes[si], nodes[di], writers[si][di]
		msg := fmt.Sprintf("ping #%d from %v to %v\n", count, src.ID(), dst.ID())
		_, err := writer.WriteString(msg)
		if err != nil {
			return err
		}
		fmt.Printf("sent random ping: %s\n", msg)
		count++
		// wait for a fixed interval
		time.Sleep(3 * time.Second)
	}
}
