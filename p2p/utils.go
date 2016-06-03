package p2p

import (
	"net"
	"strings"
)

func IsOwnIP(IP net.IP) bool {

	ifaces, err := net.Interfaces()
	// handle err

	if err != nil {
		return true
	}

	for _, i := range ifaces {
	    addrs, err := i.Addrs()
	    // handle err
	    if err != nil {
			return true
		}

	    for _, addr := range addrs {
	        var ip net.IP
	        switch v := addr.(type) {
	        case *net.IPNet:
	                ip = v.IP
	        case *net.IPAddr:
	                ip = v.IP
	        }

	        // process IP address
	        if ip.Equal(IP) {
	        	return true
	        }
	    }
	}
	return false
}

func IsPrivateIP(ip net.IP) bool {
	if !strings.HasPrefix(ip.String(), "10.") && !strings.HasPrefix(ip.String(), "192.168.") && !strings.HasPrefix(ip.String(), "172.16.") {
		return false
	} else {
		return true
	}
}
