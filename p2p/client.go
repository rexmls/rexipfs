package p2p

import (
	"time"
	"fmt"
	"io/ioutil"
	"bytes"
	"net/http"
	"encoding/json"
	"strconv"
	"github.com/rexmls/rexipfs/content"
)

type Client struct {
	ListenPort int
	DebugMode bool
}

var MyClient = new (Client)

func (c *Client) Start() {

	var keepRunning bool = true

	//just wait before starting peer connections
	time.Sleep(2 * time.Second)	

	for keepRunning == true {

		var tmpPeer *RexIpfsPeer

		//first check if we have any wanted content
		if MyWantList.Count() > 0 {
			tmpPeer = MyPeerList.GetNextPeer(0)
		} else {
			// if no content get the next peer who hasn't been used for 5 minutes
			tmpPeer = MyPeerList.GetNextPeer(300)
		}

		if tmpPeer != nil {
			c.DoSwaps(tmpPeer)
		}

		//Do some cleaning up, remove old stale peers
		MyPeerList.RemoveOldPeers()	

		time.Sleep(1 * time.Second)	
	}
}

func (c *Client) DoSwaps(peer *RexIpfsPeer) {
	fmt.Printf("Initiating swaps with peer: %v:%v...\n", peer.ip, peer.port)

	remoteSSS := c.SwapStats(peer)

	//if remote reponse was not correct then bail
	if remoteSSS == nil {
		return
	}

	//if remote has no wants and cannot provide us with any haves then bail
	if MyWantList.Count() > 0 && len(remoteSSS.Haves) == 0 {
		fmt.Println("Peer session was unproductive\n")
		peer.consecutive_nocontent++
	}

	if len(remoteSSS.Wanted) == 0 && len(remoteSSS.Haves) == 0 {
		//no need to process to swap content
		return
	}

	c.SwapContent(peer, remoteSSS)	
}

func (c *Client) SwapStats(peer *RexIpfsPeer) *SwapStatsStruct {

	var sss = new(SwapStatsStruct)

	sss.ListenPort = c.ListenPort
	sss.Peers = MyPeerList.Get10PeersForSwap(peer.ip)
	sss.Wanted = MyWantList.Get10WantsForSwap()

	jsonBytes, err := json.Marshal(sss)

	if MyClient.DebugMode {
		fmt.Printf("Sending = %s\n", jsonBytes)
	}

	buf := bytes.NewBuffer(jsonBytes)

	resp, err := http.Post("http://" + peer.ip.String() + ":" + strconv.Itoa(peer.port) + "/api/v0/swapstats", "", buf)

	peer.dial_attempts++
	peer.last_dialed = time.Now()

	if err != nil {
		fmt.Printf("Could not connect to %s\n", peer.ip.String())
		peer.last_dial_status = DIALSTATUS_ERROR
		peer.consecutive_errors++
		return nil
	}

	peer.successful_dials++
	peer.last_connected = time.Now()

	body, err := ioutil.ReadAll(resp.Body)

	if MyClient.DebugMode {
		fmt.Printf("Received = %s\n", body)
	}

	if err != nil {
		fmt.Printf("Client.SwapStats error: %v\n", err)
		peer.consecutive_errors++
		return nil
	}

	peer.consecutive_errors = 0	
	resp.Body.Close() 

	//decode the remote sides payload
	var remoteSSS = new(SwapStatsStruct)
	json.Unmarshal(body, remoteSSS)

	//add their peers
	MyPeerList.AddPeers(remoteSSS.Peers)

	return remoteSSS
}

func (c *Client) SwapContent(peer *RexIpfsPeer, sss *SwapStatsStruct) {

	if len(sss.Haves) == 0 {
		peer.consecutive_nocontent++
	} 

	var scs = new(SwapContentStruct)

	//build our content to send
	contentToSend := make([]SwapContentItemStruct, 0)
	for a := 0; a < len(sss.Wanted); a++ {
		tmpContent := new(SwapContentItemStruct)
		tmpContent.Content_b64 = content.Get(sss.Wanted[a])
		tmpContent.Hash = sss.Wanted[a]
		contentToSend = append(contentToSend, *tmpContent) 
	}

	scs.Content = contentToSend
	
	scs.Send = sss.Haves

	jsonBytes, err := json.Marshal(scs)

	if MyClient.DebugMode {
		fmt.Printf("Sending = %s\n", jsonBytes)
	}

	buf := bytes.NewBuffer(jsonBytes)

	resp, err := http.Post("http://" + peer.ip.String() + ":" + strconv.Itoa(peer.port) + "/api/v0/swapcontent", "", buf)

	peer.dial_attempts++
	peer.last_dialed = time.Now()

	if err != nil {
		fmt.Printf("Error in Client.SwapContent http.Post: %v\n", err)
		peer.last_dial_status = DIALSTATUS_ERROR
		peer.consecutive_errors++
		return
	}

	peer.successful_dials++
	peer.last_connected = time.Now()

	body, err := ioutil.ReadAll(resp.Body)

	if MyClient.DebugMode {
		fmt.Printf("Received = %s\n", body)
	}

	if err != nil {
		fmt.Printf("Error in Client.SwapContent ioutil.ReadAll: %v\n", err)
		peer.consecutive_errors++
		return
	}

	peer.consecutive_errors = 0	
	resp.Body.Close()

	//decode the remote sides payload
	var remoteSCS = new(SwapContentStruct)
	json.Unmarshal(body, remoteSCS)

	//TODO: decode b64 and add to IPFS
	//loop through remote Content provided and add it to IPFS
	for a := 0; a < len(remoteSCS.Content); a++ {
		//first check its on out wantlist
		if !MyWantList.Has(remoteSCS.Content[a].Hash) {
			continue
		}

		//then add it to IPFS
		bSuccess := content.Add(remoteSCS.Content[a].Hash, remoteSCS.Content[a].Content_b64)

		if !bSuccess {
			fmt.Println("Problem adding received content to IPFS")
			return
		}

		//then remove it from wantlist
		MyWantList.Remove(remoteSCS.Content[a].Hash)
	}

	//add received content to IPFS
	if len(remoteSCS.Content) == 0 {
		peer.consecutive_nocontent++
	}
}



