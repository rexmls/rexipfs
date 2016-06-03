package p2p

import (
	"fmt"
	"net"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/rexmls/rexipfs/content"
)

type Listener struct {}

var MyListener = new(Listener)

func (l *Listener) SwapStats(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Listener: Got SwapStats request from %v\n", req.RemoteAddr)
	body, _ := ioutil.ReadAll(req.Body)

	if MyClient.DebugMode {
		fmt.Printf("Received = %s\n", body)
	}

	var remoteSSS = new (SwapStatsStruct)
	var ourSSS = new(SwapStatsStruct)

	json.Unmarshal(body, remoteSSS)

	//add the connecting peer to our peer list (using the listen port provided) return the new peer object
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	hostIP := net.ParseIP(host)
	MyPeerList.Add(hostIP, remoteSSS.ListenPort)

	peer := MyPeerList.GetPeer(hostIP)
	peer.last_connected = time.Now()
	peer.incoming_attempts++

	//add the provided peers to our peer list
	MyPeerList.AddPeers(remoteSSS.Peers)

	//loop through remote want list and return hashes we have
	var haves = make([]string, 0)
	for a := 0; a < len(remoteSSS.Wanted); a++ {
		if content.HaveContent(remoteSSS.Wanted[a]) {
			haves = append(haves, remoteSSS.Wanted[a])
		}
	}

	ourSSS.ListenPort = MyClient.ListenPort

	//check if incoming connection is private
	ourSSS.Peers = MyPeerList.Get10PeersForSwap(peer.ip)
	ourSSS.Wanted = MyWantList.Get10WantsForSwap()
	ourSSS.Haves = haves

	outBody, _ := json.Marshal(ourSSS)

	if MyClient.DebugMode {
		fmt.Printf("Sent = %s\n", outBody)
	}

	rw.Write(outBody)
}

func (l *Listener) SwapContent(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Listener: Got SwapContent request from %v\n", req.RemoteAddr)
	body, _ := ioutil.ReadAll(req.Body)

	if MyClient.DebugMode {
		fmt.Printf("Received = %s\n", body)
	}

	var remoteSCS = new (SwapContentStruct)
	var ourSCS = new(SwapContentStruct)

	json.Unmarshal(body, remoteSCS)

	//get the peer object
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	hostIP := net.ParseIP(host)

	peer := MyPeerList.GetPeer(hostIP)
	peer.last_connected = time.Now()

	//loop through remote Content provided and add it to IPFS
	for a := 0; a < len(remoteSCS.Content); a++ {
		//first check its on out wantlist
		if !MyWantList.Has(remoteSCS.Content[a].Hash) {
			continue
		}

		//then add it to IPFS
		content.Add(remoteSCS.Content[a].Hash, remoteSCS.Content[a].Content_b64)

		//then remove it from wantlist
		MyWantList.Remove(remoteSCS.Content[a].Hash)
	}

	//loop through Sends requested and send the base64 data
	contentToSend := make([]SwapContentItemStruct, 0)
	for a := 0; a < len(remoteSCS.Send); a++ {
		tmpContent := new(SwapContentItemStruct)
		tmpContent.Content_b64 = content.Get(remoteSCS.Send[a])
		tmpContent.Hash = remoteSCS.Send[a]
		contentToSend = append(contentToSend, *tmpContent) 
	}

	ourSCS.Content = contentToSend

	outBody, _ := json.Marshal(ourSCS)

	if MyClient.DebugMode {
		fmt.Printf("Sent = %s\n", outBody)
	}

	rw.Write(outBody)
}
