package p2p

import (
    "time"
    "net"
    //"fmt"
)

const (
    DIALSTATUS_NOTTRIED = iota
    DIALSTATUS_ERROR
    DIALSTATUS_SUCCESS
)

type RexIpfsPeer struct {

    ip net.IP
    port int

    //is_private bool
    blocked bool

    first_added time.Time
    last_connected time.Time

	last_dial_status int
	last_dialed time.Time

	dial_attempts int
	incoming_attempts int
    successful_dials int

    consecutive_errors int
    consecutive_nocontent int

	last_incoming time.Time

    last_content_provided time.Time

    average_latency [10]int
    content_provided_count int
}


func (p *RexIpfsPeer) GetWaitTime() (time.Time, int) {
    var waitTime int

    if p.consecutive_errors > 0 {
        waitTime = p.consecutive_errors * 5 // 5 seconds
        return p.GetLastAction(), waitTime
    }

    if p.consecutive_nocontent > 0 {
        waitTime = p.consecutive_nocontent * 5 // 5 seconds
        if waitTime > 600 {
            waitTime = 600 //max 10 minutes
        }
        return p.GetLastAction(), waitTime
    }

    return p.GetLastAction(), 10
}

func (p *RexIpfsPeer) GetLastAction() time.Time {

    var lastAction time.Time
    if p.last_dialed.After(lastAction) {
        lastAction = p.last_dialed
    }

    if p.last_incoming.After(lastAction) {
        lastAction = p.last_incoming
    }

    if p.last_connected.After(lastAction) {
        lastAction = p.last_connected
    }


    return lastAction
}




