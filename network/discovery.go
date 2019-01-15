package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	connectedToDirectory = false
	peers                map[string]string
	peersMapChannel      chan map[string]string
	chatPort             int
	usernameChannel      chan string
)

// ConnectToDirectory ...
func ConnectToDirectory(directoryServer string, directoryPort int, localChatPort int, peersMap chan map[string]string, usernameChan chan string) {
	peersMapChannel = peersMap
	peers = make(map[string]string)
	chatPort = localChatPort
	usernameChannel = usernameChan

	conn, err := net.Dial("tcp", directoryServer+":"+strconv.Itoa(directoryPort))
	if err != nil {
		fmt.Println("Cannot connect to directory", err)
		return
	}

	fmt.Println("Connected to directory.")

	go listenFromDirectory(conn)
	connectedToDirectory = true
	send(conn, "HELLO", strconv.Itoa(chatPort))

	for {
		if !connectedToDirectory {
			conn, err := net.Dial("tcp", directoryServer+":"+strconv.Itoa(directoryPort))
			if err != nil {
				fmt.Println("Cannot connect to directory", err)
			} else {
				connectedToDirectory = true
				send(conn, "HELLO", strconv.Itoa(chatPort))
				go listenFromDirectory(conn)
			}

		}

		time.Sleep(2 * time.Second)
	}
}

func listenFromDirectory(conn net.Conn) {
	for {
		message, err := GetNextMessage(conn)
		if err != nil {
			fmt.Println("Lost connection to directory ", err)
			connectedToDirectory = false
			return
		}

		//fmt.Println(conn.RemoteAddr(), "said :", message)

		switch message.Kind {
		case "PEERS":
			handlePeers(message.Data)

		case "NAME":
			handleName(message.Data)

		case "WELCOME":
			handleWelcome(message.Data)

		default:
			fmt.Println("Unknown message kind :", message)
		}
	}
}

func handlePeers(sList string) {
	list := strings.Split(sList, " ")

	newPeersList := make(map[string]string)

	// convert peers list to a map to simplify search by address
	// (and we only keep valid addresses)
	for _, addr := range list {
		if len(strings.SplitN(addr, ":", 2)) != 2 {
			continue
		}
		if _, found := peers[addr]; found {
			newPeersList[addr] = peers[addr]
		} else {
			newPeersList[addr] = addr
			peers[addr] = newPeersList[addr]
		}
	}

	peersMapChannel <- newPeersList
	peers = newPeersList
}

func handleName(data string) {
	list := strings.SplitN(data, " ", 2)
	if len(list) < 2 {
		fmt.Println("Invalid NAME message", data)
		return
	}

	addr, newName := strings.TrimSpace(list[0]), strings.TrimSpace(list[1])

	peers[addr] = newName

	fmt.Println(addr, "is now", peers[addr])

	peersMapChannel <- peers
}

func handleWelcome(data string) {
	if len(strings.Split(data, " ")) != 1 {
		fmt.Println("Invalid WELCOME message", data)
		return
	}

	usernameChannel <- data
}

func send(conn net.Conn, msgType string, data string) (int, error) {
	return conn.Write([]byte(msgType + " " + data + "\n"))
}
