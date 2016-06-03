package p2p

import (
	//"fmt"
	"time"
	"net"
	"strconv"
)

type aPeerList struct {
	peers map[string]*RexIpfsPeer
	//peersSorted []string
}

func NewPeerList() *aPeerList {
	return &aPeerList{
		peers: make(map[string]*RexIpfsPeer, 0),
		//peersSorted: make([]string, 0),
	}
}

var MyPeerList = NewPeerList()

func (pl *aPeerList) Add(IP net.IP, port int) {

	//check if the peer already exists
	_, ok := pl.peers[IP.String()]
	if ok {
		return
	}

	//check if peer is ourselves
	if IsOwnIP(IP) {
		return
	}

	tmpPeer := new(RexIpfsPeer)

	tmpPeer.ip = IP
	tmpPeer.port = port
	tmpPeer.first_added = time.Now()
	//tmpPeer.dial_attempts = 0

	pl.peers[IP.String()] = tmpPeer
	//peersSorted
}

func (pl *aPeerList) AddPeers(peers []string) {
	for i := 0; i < len(peers); i++ {
		host, port, _ := net.SplitHostPort(peers[i])
		hostIP := net.ParseIP(host)
		iPort, _ := strconv.Atoi(port)
		pl.Add(hostIP, iPort)
	}
}

func (pl *aPeerList) Remove(address string) {
	delete(pl.peers, address)
}

func (pl *aPeerList) GetPeer(ip net.IP) *RexIpfsPeer {
	return pl.peers[ip.String()]
}

func (pl *aPeerList) Count() int {
	return len(pl.peers)
}

func (pl *aPeerList) GetNextPeer(lastActionSecondsAgo int) *RexIpfsPeer {

	var bestPeer *RexIpfsPeer

	for _, peer := range pl.peers {

		//get the last action time and waitime
		lastActionTime, waitTime := peer.GetWaitTime()

		//check lastActionSecondsAgo
		//fmt.Printf("%v\n", time.Since(lastActionTime).Seconds())
		if time.Since(lastActionTime).Seconds() < float64(lastActionSecondsAgo) {
			continue
		}

		//if peer has never been connected to or been connected from
		if peer.dial_attempts == 0 && peer.incoming_attempts == 0 {
			bestPeer = peer
			break
		}

		// check if the waittime has expired
		
		if time.Since(lastActionTime).Seconds() > float64(waitTime) {
			bestPeer = peer
			break
		}
	}

	return bestPeer
}

func (pl *aPeerList) RemoveOldPeers() {
	for address, peer := range pl.peers {

		if peer.last_connected.IsZero() && time.Since(peer.first_added).Hours() > 336 {
			pl.Remove(address)
			continue
		}

		if peer.last_connected.IsZero() == false && time.Since(peer.last_connected).Hours() > 336 { // 14 days
			pl.Remove(address)
			continue
		}
	}
}

func (pl *aPeerList) Get10PeersForSwap(forIP net.IP) []string {

	forPrivate := IsPrivateIP(forIP)

	i := 0
	keys := make([]string, 0)
	for k, p := range pl.peers {

		//skip private/public peers
		if !forPrivate && IsPrivateIP(p.ip) {
			continue
		}

		//dont send peer his own address
		if forIP.String() == k {
			continue
		}


	    keys = append(keys, k + ":" + strconv.Itoa(p.port))
	    i++
	    if i >= 10 {
	    	break
	    }
	}
	return keys
}

