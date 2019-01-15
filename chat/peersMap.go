package chat

import (
	"strconv"
	"strings"

	"github.com/teanan/GOssip-TP/network"
)

// peersMap is a map of Peers identified by their full address ("a.b.c.d:0000")
// peersMap.localUsername is used to store the username of the local client
type peersMap struct {
	peers         map[string]network.Peer
	localUsername string
}

// Get returns the peer identified by its full address ("a.b.c.d:0000")
func (pmap *peersMap) Get(addr string) network.Peer {
	return pmap.peers[addr]
}

// Set updates the peer identified by its full address ("a.b.c.d:0000")
func (pmap *peersMap) Set(addr string, peer network.Peer) {
	// Question 1
}

// Find looks for a peer identified by its full address ("a.b.c.d:0000")
// first return parameter is true if we found it, false otherwise
func (pmap *peersMap) Find(address string) (bool, network.Peer) {
	// Question 1
	return false, network.Peer{}
}

// FindByName looks for a peer identified by its username ("my_user_name")
// first return parameter is true if we found it, false otherwise
func (pmap *peersMap) FindByName(name string) (bool, network.Peer) {
	// Question 3
	return false, network.Peer{}
}

// SendToAll adds a network.Message to the sending queue of every known peer
func (pmap *peersMap) SendToAll(msg network.Message) {
	// Question 1
}

// SendToAll adds a network.Message to the sending queue of said peer
func (pmap *peersMap) SendTo(peer network.Peer, msg network.Message) {
	peer.Send <- msg
}

// SetNewPeersList updates the known peers map with newly received list from the directory server
// execute the callbacks onPeerConnected (onPeerDisconnected) when a new peer is connected (disconnected)
func (pmap *peersMap) SetNewPeersList(newList map[string]string, onPeerConnected func(network.Peer), onPeerDisconnected func(network.Peer)) {
	// remove peers that are no longer present
	for addr := range pmap.peers {
		_, found := newList[addr]
		if !found {
			onPeerDisconnected(pmap.peers[addr])
			delete(pmap.peers, addr)
		}
	}

	// add new peers
	for addr := range newList {
		_, found := pmap.peers[addr]
		if found {
			continue
		}

		port, _ := strconv.Atoi(strings.SplitN(addr, ":", 2)[1])
		peer := network.CreatePeer(
			strings.SplitN(addr, ":", 2)[0],
			port,
		)
		pmap.peers[addr] = peer
		onPeerConnected(peer)
	}
}

// SetLocalUsername set the username of local client
func (pmap *peersMap) SetLocalUsername(localUsername string) {
	pmap.localUsername = localUsername
}

// GetLocalUsername returns the username of local client
func (pmap *peersMap) GetLocalUsername() string {
	return pmap.localUsername
}

// NewPeersMap builds a new empty peersMap
func NewPeersMap() *peersMap {
	return &peersMap{
		peers: make(map[string]network.Peer),
	}
}
