package network

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// Dial connects to a Peer and is in charge of sending outgoing messages to this peer.
// localChatPort is our own listening port and is used so the other peer can recognise us.
func Dial(peer Peer, localChatPort int) {
	conn, err := net.Dial("tcp", peer.address+":"+strconv.Itoa(peer.port))
	if err != nil {
		fmt.Println("Failed to connect to peer", err)
		return
	}

	fmt.Println("Connected to", peer.address+":"+strconv.Itoa(peer.port))

	// Sending a HELLO message with our local port
	// this acts as authentication between peers
	Message{
		"HELLO",
		strconv.Itoa(localChatPort),
	}.Send(conn)

	for {
		select {
		case msg := <-peer.Send:
			msg.Send(conn)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
