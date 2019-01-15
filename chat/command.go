package chat

import (
	"fmt"
	"strings"

	"github.com/teanan/GOssip-TP/network"
)

// commandProcessor handles outgoing messages to other peers
// it contains a pointer to the common peersMap and a channel to output to the screen
type commandProcessor struct {
	peers         *peersMap
	messageOutput chan<- string
}

// Process handles raw text messages from the command line or webui
func (processor *commandProcessor) Process(command string) {

	if strings.HasPrefix(command, "/") {
		commandName := ""

		// question 5

		switch commandName {
		default:
			fmt.Print("Unknown command", commandName)
		}
	} else {
		processor.say(command)
	}
}

// say sends outgoing messages of kind SAY
func (processor *commandProcessor) say(command string) {
	processor.messageOutput <- "[" + processor.peers.GetLocalUsername() + "] " + command
	processor.peers.SendToAll(network.Message{
		Kind: "SAY",
		Data: command,
	})
}

// say sends outgoing messages of kind SAYTO (private messages)
func (processor *commandProcessor) sayTo(commandParams string) {

}

// NewCommandProcessor builds a new CommandProcessor with pointer to the common peersMap and channel to output to the screen
func NewCommandProcessor(peers *peersMap, messageOutput chan<- string) *commandProcessor {
	return &commandProcessor{
		peers:         peers,
		messageOutput: messageOutput,
	}
}
