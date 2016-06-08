package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"sync"
	"strings"
	"github.com/rexmls/rexipfs/content"
	"github.com/rexmls/rexipfs/p2p"
	"strconv"
)

var defaultIPFSAPIListenAddress = "127.0.0.1:5002"
var defaultP2PListenAddress = "0.0.0.0:7400"
var defaultUpstreamIPFS = "127.0.0.1:5001"

var argIPFSAPIListenAddress *string
var argP2PListenAddress *string
var argUpstreamIPFS *string

func init() {

	argIPFSAPIListenAddress = flag.String("client_ipfs_address", defaultIPFSAPIListenAddress, "Address:Port for the internal IFPS compatible API to listen on")
	argP2PListenAddress = flag.String("p2p_address", defaultP2PListenAddress, "Address:Port for the public p2p API to listen on")
	argUpstreamIPFS = flag.String("upstream_ipfs", defaultUpstreamIPFS, "Address:Port for the IPFS node to store and retrieve the content")

	var argPeers = flag.String("peers", "", "IP addresses of additional peers to bootstrap from. Delimited with comma and no spaces")
	var argDebugMode = flag.Bool("debug", false, "Debug mode with more logging to console")

	flag.Parse()

	str_argUpstreamIPFS := *argUpstreamIPFS
	if !strings.HasPrefix(str_argUpstreamIPFS, "http://") {
		str_argUpstreamIPFS = "http://" + str_argUpstreamIPFS
	}
	content.UpstreamIPFSAddress = str_argUpstreamIPFS

	_, port, _ := net.SplitHostPort(*argP2PListenAddress)
	iPort, _ := strconv.Atoi(port)
	p2p.MyClient.ListenPort = iPort
	p2p.MyClient.DebugMode = *argDebugMode

	if len(*argPeers) > 0 {
		tmpPeerArray := strings.Split(*argPeers,",")
		for a := 0; a < len(tmpPeerArray); a++ {
			hostIP, port, _ := net.SplitHostPort(tmpPeerArray[a])
			iPort, _ := strconv.Atoi(port)
			p2p.MyPeerList.Add(net.ParseIP(hostIP), iPort)	
		}
	}
}

func handleIPFSRequests (rw http.ResponseWriter, req *http.Request) {
	multihash, bFound := content.HttpObjectGet(rw, req)
	if !bFound {
		p2p.MyWantList.Add(multihash)
	}
}

func apiCatchall(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Unknown request")
	fmt.Printf("%s", req.URL)
}

func main() {

	fmt.Printf("DebugMode: %v\n", p2p.MyClient.DebugMode)
	fmt.Printf("IPFS Node: %s\n", content.UpstreamIPFSAddress)

	serverMuxA := http.NewServeMux()
	serverMuxA.HandleFunc("/api/v0/swapstats", p2p.MyListener.SwapStats)
	serverMuxA.HandleFunc("/api/v0/swapcontent", p2p.MyListener.SwapContent)
	serverMuxA.HandleFunc("/", apiCatchall)

	serverMuxB := http.NewServeMux()
	serverMuxB.HandleFunc("/api/v0/add", content.HttpAdd)
	serverMuxB.HandleFunc("/api/v0/object/get", handleIPFSRequests)
	serverMuxB.HandleFunc("/api/v0/cat", handleIPFSRequests)
	serverMuxB.HandleFunc("/shutdown", content.HttpShutdown)
	serverMuxB.HandleFunc("/", apiCatchall)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		//this is a IPFS compatible API only accessible locally
		fmt.Printf("Listening on %v\n", *argIPFSAPIListenAddress)
		http.ListenAndServe(*argIPFSAPIListenAddress, serverMuxB)
		wg.Done()
	}()

	go func() {
		// this is a custom protocol for exchanging content outside of IPFS
		fmt.Printf("Listening on %v\n", *argP2PListenAddress)
		http.ListenAndServe(*argP2PListenAddress, serverMuxA)
		wg.Done()
	}()

	p2p.MyClient.Start()

	wg.Wait()
}
