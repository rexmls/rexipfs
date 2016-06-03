package p2p

import (
	//"fmt"
	"time"
	//"github.com/rexmls/rexipfs/content"
	"sync"
)

type aWantList struct {
	sync.RWMutex
	wants map[string]*wantItem
}

func NewWantList() *aWantList {
	return &aWantList{wants: make(map[string]*wantItem, 0)}
}

var MyWantList = NewWantList()

type wantItem struct {
	first_added time.Time
	last_checked time.Time
	check_count int
}

func (wl *aWantList) Add(hash string) {

	_, ok := wl.wants[hash]

	//if its already in the wantlist ignore it
	if ok {
		return
	}

	//if its already in our IPFS ignore it
	//if content.HaveContent(hash) {
	//	fmt.Printf("Already have %s\n", hash)
	//	return
	//}

	tmpWantItem := new(wantItem)

	tmpWantItem.first_added = time.Now()
	tmpWantItem.check_count = 0

	wl.Lock()
	wl.wants[hash] = tmpWantItem
	wl.Unlock()
}

func (wl *aWantList) Remove(hash string) {
	wl.Lock()
	delete(wl.wants, hash)
	wl.Unlock()
}

func (wl *aWantList) Count() int {
	wl.RLock()
	count := len(wl.wants)
	wl.RUnlock()

	return count
}

func (wl *aWantList) Has(hash string) bool {
	wl.RLock()
	_, ok := wl.wants[hash]
	wl.RUnlock()
	if ok {
		return true
	} else {
		return false
	}
}

func (wl *aWantList) Get10WantsForSwap() []string {
	i := 0
	wl.RLock()
	keys := make([]string, len(wl.wants))
	for k := range wl.wants {
	    keys[i] = k
	    i++
	    if i >= 10 {
	    	break
	    }
	}
	wl.RUnlock()
	return keys
}
