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
	"github.com/libp2p/go-libp2p/core/protocol"
)

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
	fmt.Printf("created nodes: %v\n", nodes)

	streams, err := openConnectionStreams(nodes)
	if err != nil {
		panic(err)
	}
	fmt.Printf("opened streams: %v\n", streams)

	listenForPings(nodes)
	fmt.Printf("listening for pings...")

	err = sendRandomPings(nodes, streams)
	if err != nil {
		panic(err)
	}
	fmt.Printf("sent random pings. terminating...")
}

// create N nodes
func createNodes(n int) ([]host.Host, error) {
	nodes := make([]host.Host, 0, n)
	for i := 0; i < n; i++ {
		node, err := libp2p.New()
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// establish connections between all pairs of nodes & open streams for ping protocol
func openConnectionStreams(nodes []host.Host) ([][]network.Stream, error) {
	streams := make([][]network.Stream, len(nodes))
	// between every pair of nodes
	for si, src := range nodes {
		streams[si] = make([]network.Stream, len(nodes))
		for di, dst := range nodes {
			if si == di || streams[si][di] != nil {
				continue
			}
			// establish connection
			if err := src.Connect(context.Background(), *host.InfoFromHost(dst)); err != nil {
				return streams, err
			}
			// open stream
			stream, err := src.NewStream(context.Background(), dst.ID(), ProtocolPing)
			if err != nil {
				return streams, err
			}
			// store bidirectional stream
			streams[si][di] = stream
			streams[di][si] = stream
		}
	}
	return streams, nil
}

// listen for pings on all nodes & display received pings
func listenForPings(nodes []host.Host) {
	// on all nodes
	for _, node := range nodes {
		// listen for pings
		node.SetStreamHandler(ProtocolPing, func(stream network.Stream) {
			// read pings
			line, err := bufio.NewReader(stream).ReadString('\n')
			if err != nil {
				panic(err)
			}
			// display pings
			fmt.Printf("received [%s]", line)
		})
	}
}

// send ping from one random node to another indefinitely
func sendRandomPings(nodes []host.Host, streams [][]network.Stream) error {
	for i := 1; ; {
		// between random pair of nodes
		src := rand.Intn(len(nodes))
		dst := rand.Intn(len(nodes))
		if src == dst {
			continue
		}
		// send ping
		stream := streams[src][dst]
		msg := []byte(fmt.Sprintf("ping #%d from %s to %s\n", i, nodes[src].ID(), nodes[dst].ID()))
		_, err := stream.Write(msg)
		if err != nil {
			return err
		}
		// wait for a fixed interval
		time.Sleep(3 * time.Second)
		i++
	}
}
