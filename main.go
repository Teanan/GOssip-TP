package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/teanan/GOssip-TP/chat"
	"github.com/teanan/GOssip-TP/network"
)

var (
	chatPort        int           // local port for incoming chat messages
	directoryPort   = 8080        // port of the directory server to connect to
	directoryServer = "127.0.0.1" // ip of the directory server to connect to

	messageOutputChannel = make(chan string, 5) // queue of messages to print on the local screen
)

func main() {
	fmt.Println("== GOssip ==")

	// Selecting a random local port
	rand.Seed(time.Now().UnixNano())
	chatPort = 9000 + rand.Intn(1000)

	// If the program as arguments, read a new directoryServer IP
	if len(os.Args) > 1 {
		directoryServer = os.Args[1]
	}

	fmt.Println("Listening on port", chatPort)

	/* Question 4
	browserPort := 13000 + rand.Intn(1000)
	webpage, err := browser.Connect("localhost", browserPort)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("Successfuly connected to browser webpage")
	*/

	// Create channels to receive a new list of peers addresses, and a new username
	peersListChannel := make(chan map[string]string, 5)
	usernameChannel := make(chan string, 5)

	// Create a new PeersMap to store known peers
	peersMap := chat.NewPeersMap()

	// Create CommandProcessor and MessageReceiver to handle outgoing and incoming messages
	commandProcessor := chat.NewCommandProcessor(peersMap, messageOutputChannel)
	messageReceiver := chat.NewMessageReceiver(peersMap, messageOutputChannel)

	// Start listening for incoming peers connections
	go network.Listen(chatPort, peersMap, messageReceiver)

	// Start connection to the peers directory (which will send us the list of other peers)
	go network.ConnectToDirectory(directoryServer, directoryPort, chatPort, peersListChannel, usernameChannel)

	// Start reading text from the command line
	stdin := make(chan string)
	go readStdin(stdin)

	/* Question 4
	loop:
	*/
	for {

		select {

		case text, ok := <-stdin: // New command from stdin

			if !ok {
				return
			}

			commandProcessor.Process(text)

		case newList := <-peersListChannel: // New peers list from discovery server
			peersMap.SetNewPeersList(newList, onPeerConnected, onPeerDisconnected)

		case name := <-usernameChannel: // Assigned username from discovery server
			peersMap.SetLocalUsername(name)

		case message := <-messageOutputChannel: // New message to print on the screen
			fmt.Println(message)

			/* Question 4
			case <-webpage.Disconnected:
				fmt.Println("Browser webpage has disconnected")
				break loop
			*/
		}
	}
}

func onPeerConnected(peer network.Peer) {
	// Question 2
}

func onPeerDisconnected(peer network.Peer) {
	// Question 2
}

// Routine reading text from the command line
func readStdin(ch chan string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			close(ch)
			return
		}
		ch <- s
	}
}
