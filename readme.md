RexIPFS
=======

A replacement P2P library for IPFS that requires zero changes to upstream IPFS API daemons and downstream IPFS API clients.

It works by sitting as a 'man-in-the-middle' in between the client and IPFS daemon accepting IPFS API requests and forwarding them on to the IPFS daemon.  If the daemon does not have the content RexIPFS maintains its own wantlist and attempts to obtain the content via its own simple p2p network.

### Usage
```
git clone github.com/rexmls/rexipfs
go build
./rexipfs
```

For help: `./rexipfs -h`

### Why?
The current implementation of IPFS uses a lot of RAM and Network resources. I believe mainly due to the large number of peers it connects to and the chatty libP2P library.

In my tests a 512Mb RAM VPS with 500GB of monthly data was not sufficient to run a swarm connected IPFS daemon.

### Future plans
RexIPFS is (hopefully) a temporary fix, it still uses IPFS at the backend to store and retrieve content.  Once sufficient controls are included in the official IPFS implementation to limit RAM and bandwidth usage then it's simply a matter of shutting down RexIPFS, point your client back to the default http://127.0.0.1:5001 (or whatever you use).

#### Why did I not contribute to IPFS directly?
I tried tackling the bandwidth issues and peer handling in IPFS, unfortunately the IPFS/libP2P codebase takes some time to comprehend and I needed something quickly. RexIPFS is the result.