package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"golang.org/x/xerrors"
	"log"
	"math/rand"
	"os"
	"time"
)

// TODO context.WithCancel()

// TODO use zerologger: logger = logger.String("node", h.Id())
// TODO host.Host doc
// TODO network.Stream doc
// TODO bufio.ReadWriter and io.ReadWriter doc

// TODO why formatting Peer.Id() type Id string with %s does not call String() like %v?
// TODO context.WithCancel() purpose, needed?
// TODO example echo host: tests

const (
	protocolPing  = protocol.ID("my/ping/1.0.0")
	totalHosts    = 2
	pingDelimiter = '\n'
)

// set up N nodes to send random pings to each other & display received pings
func main() {
	runCommand()
	os.Exit(0)
}

func runCommand() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	create := flag.Bool("create", false, "Create a new host")
	connect := flag.Bool("connect", false, "Connect to other hosts")
	flag.Parse()

	if *create {
		runCreate()
	} else if *connect {
		runConnect()
	}
}

func runCreate() {
	h, err := createHost()
	if err != nil {
		panic(err)
	}
	defer h.Close()

	err = storeIdentity(h)
	if err != nil {
		panic(err)
	}

	showAddrs(h)
}

func runConnect() {
	ids, addrsByPeer, err := processArgs(flag.Args())
	if err != nil {
		panic(err)
	}

	h, err := startHost(ids[0], addrsByPeer[0]...)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	err = addPeers(h, ids[1:], addrsByPeer[1:])
	if err != nil {
		panic(err)
	}

	connectPeers(h, protocolPing, pingOnce)
	select {}
}

func runAuto() {
	hosts, err := createHosts(totalHosts)
	if err != nil {
		panic(err)
	}
	// listen for incoming streams (as listeners)
	startListeners(hosts, protocolPing, func(in network.Stream) {
		src, dst := in.Conn().RemotePeer(), in.Conn().LocalPeer()
		log.Printf("%v accepted an incoming stream from %v\n", dst, src)
		// on listener's side of stream
		runPingProtocol(len(hosts), in)
	})
	fillPeerStores(hosts)
	// initiate outgoing streams (as initiators)
	out, err := openStreams(hosts, protocolPing)
	if err != nil {
		panic(err)
	}
	// on initiators' side of streams
	runPingProtocol(len(hosts), out...)

	select {}
}

// create n nodes
func createHosts(n int) ([]host.Host, error) {
	fmt.Printf("creating nodes...\n")

	nodes := make([]host.Host, 0, n)
	for i := 0; i < n; i++ {
		node, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
		if err != nil {
			return nodes, xerrors.Errorf("could not start host: %v", err)
		}
		nodes = append(nodes, node)
		fmt.Printf("created %d/%d nodes: %v\n", i+1, n, formatMultiAddrs(node))
	}

	return nodes, nil
}

// start listening for given protocol indefinitely on all nodes
func startListeners(nodes []host.Host, pid protocol.ID, handler network.StreamHandler) {
	log.Printf("starting to listen for protocol %v...", pid)

	for i, node := range nodes {
		// handler called as many times as the number of incoming streams
		node.SetStreamHandler(protocolPing, handler)
		log.Printf("started listening on %d/%d nodes...\n", i+1, len(nodes))
	}
}

// add addresses of all peers to each node's peer store
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
		fmt.Printf("filled %d/%d peer stores: {node: %v, peer store: %s}\n", i+1, len(nodes), node.ID(), formatPeers(node.Peerstore()))
	}
}

// open outgoing streams for given protocol (and connections if necessary) between all pairs of nodes
func openStreams(nodes []host.Host, p protocol.ID) ([]network.Stream, error) {
	log.Printf("opening streams for protocol %v...\n", p)

	totalStreams := len(nodes) * (len(nodes) - 1) / 2
	streams := make([]network.Stream, 0, totalStreams)
	n := 0
	// between every pair of nodes
	for si, src := range nodes {
		for di, dst := range nodes {
			// don't open stream to self or if already opened by the other side
			if si >= di {
				continue
			}
			// open stream (implicitly create connection if not exist)
			s, err := src.NewStream(context.Background(), dst.ID(), p)
			if err != nil {
				return nil, xerrors.Errorf("could not open stream: %v", err)
			}
			streams = append(streams, s)
			n++
			log.Printf("opened %d/%d outgoing streams\n", n, totalStreams)
		}
	}

	return streams, nil
}

// run ping protocol over given stream(s)
func runPingProtocol(maxInterval int, streams ...network.Stream) {
	for _, s := range streams {
		// send random pings
		go sendPings(s, maxInterval)
		// display received pings
		go readPings(s)
	}
}

// send pings indefinitely at random intervals
func sendPings(s network.Stream, maxInterval int) {
	src, dst := s.Conn().LocalPeer(), s.Conn().RemotePeer()
	log.Printf("%v sending pings to %v at random intervals...", src, dst)

	// write indefinitely
	for n := 1; ; n++ {
		msg := createPing(src, dst, n)
		_, err := s.Write([]byte(msg))
		if err != nil {
			log.Fatalf("could not write to stream: %v", err)
			return
		}
		log.Printf("%v sent ping to %v: %s", src, dst, msg)
		// wait for random interval
		time.Sleep(time.Duration(rand.Intn(maxInterval)) * time.Second)
	}
}

func createPing(src, dst peer.ID, id int) string {
	// trailing '\n' is important for read to happen successfully
	return fmt.Sprintf("ping %s%s#%d%c", src[len(src)-2:], dst[len(dst)-2:], id, pingDelimiter)
}

// read and display pings indefinitely
func readPings(s network.Stream) {
	src, dst := s.Conn().RemotePeer(), s.Conn().LocalPeer()
	log.Printf("%v reading pings from %v...", dst, src)

	reader := bufio.NewReader(s)
	// read indefinitely
	for {
		msg, err := reader.ReadString(pingDelimiter)
		if err != nil {
			log.Fatalf("could not read stream buffer: %v", err)
			return
		}
		// display pings
		log.Printf("%v received ping from %v: %s", dst, src, msg)
	}
}
