package chat

import (
	"fmt"
	"strings"

	"github.com/teanan/GOssip-TP/network"
)

// MessageReceiver handles incoming messages from other peers
// it contains a pointer to the common peersMap and a channel to output to the screen
type MessageReceiver struct {
	peers         *peersMap
	messageOutput chan<- string
}

// Receive handles arriving unsorted messages
// from is the Peer who sent it
func (receiver *MessageReceiver) Receive(message network.Message, from network.Peer) {
	switch message.Kind {
	case "SAY":
		receiver.handleSay(message.Data, from)
	case "SAYTO":
		receiver.handleSayTo(message.Data, from)
	case "NAME":
		receiver.handleName(message.Data, from)
	default:
		fmt.Println("Unknown message kind :", message)
	}
}

// HandleHello is a special message used by peers to identify with each other (implements network.MessageReceiver interface)
// we use it to send our local username to the newly connected peer
func (receiver *MessageReceiver) HandleHello(data string, from network.Peer) {
	receiver.peers.SendTo(from, network.Message{
		Kind: "NAME",
		Data: receiver.peers.GetLocalUsername(),
	})
}

// handleSay is called when a message of kind "SAY" is received
// data is the value of the received message, from is the Peer who sent it
func (receiver *MessageReceiver) handleSay(data string, from network.Peer) {
	receiver.messageOutput <- fmt.Sprint("[", from, "] ", data)
}

// handleSay is called when a message of kind "SAYTO" is received
// data is the value of the received message, from is the Peer who sent it
func (receiver *MessageReceiver) handleSayTo(data string, from network.Peer) {
	// question 5
}

// handleSay is called when a message of kind "NAME" is received
// data is the value of the received message, from is the Peer who sent it
func (receiver *MessageReceiver) handleName(data string, from network.Peer) {
	// Check if the submitted name is valid
	if strings.ContainsAny(data, "\t\r\n ") {
		return
	}

	// Check if the submitted name is different from the previous one
	if data == from.String() {
		return
	}

	// Check if the submitted name is different from other peers and our own
	if found, _ := receiver.peers.FindByName(data); found || receiver.peers.GetLocalUsername() == data {
		receiver.messageOutput <- fmt.Sprint(from.String(), "tried to use an already taken username")
		return
	}

	receiver.messageOutput <- fmt.Sprint(from.String(), " is now known as ", data)
	from.SetName(data)
	receiver.peers.Set(from.FullAddress(), from)
}

// NewMessageReceiver builds a new MessageReceiver with pointer to the common peersMap and channel to output to the screen
func NewMessageReceiver(peers *peersMap, messageOutput chan<- string) *MessageReceiver {
	return &MessageReceiver{
		peers:         peers,
		messageOutput: messageOutput,
	}
}
