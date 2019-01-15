package network

import (
	"strconv"
)

// Peer represent a known peer with its address ("a.b.c.d") and port (0000).
// Send is the queue of outgoing messages to this peer
type Peer struct {
	address string
	port    int
	Send    chan Message
}

// PeersMap is an interface to a collection of Peers with Get and Find methods.
// It is used by fonctions of the network package.
// (see that it is implemented in the chat package, but we never need to import it)
type PeersMap interface {
	Get(address string) Peer
	Find(address string) (bool, Peer)
}

// String return a string version of current peer
func (p Peer) String() string {
	return p.address + ":" + strconv.Itoa(p.port)
}

// FullAddress return the full address ("a.b.c.d:0000") of current peer
func (p Peer) FullAddress() string {
	return p.address + ":" + strconv.Itoa(p.port)
}

// SetName sets the username of current peer
func (p *Peer) SetName(name string) {

}

// CreatePeer return a new Peer with said addresse and port, and a new messages queue
func CreatePeer(addr string, port int) Peer {
	return Peer{
		address: addr,
		port:    port,
		Send:    make(chan Message),
	}
}
