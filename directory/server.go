package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/teanan/GOssip-TP/network"
)

type Peer struct {
	conn        net.Conn
	address     string
	pseudo      string
	chatPort    int
	chatAddress string
}

var (
	guestNum int
	peers    = make(map[string]Peer)
)

func main() {
	fmt.Println("GOssip peers directory server")
	fmt.Println("===")

	listen(8080)
}

func listen(port int) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Failed to open listen connection", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Failed to accept incoming connection", err)
		} else {
			guestNum = guestNum + 1
			peers[conn.RemoteAddr().String()] = Peer{
				conn:        conn,
				address:     conn.RemoteAddr().String(),
				pseudo:      "Guest#" + strconv.Itoa(guestNum),
				chatPort:    0,
				chatAddress: "?",
			}

			go handleConnection(peers[conn.RemoteAddr().String()])
		}
	}
}

func handleConnection(peer Peer) {
	conn := peer.conn

	if _, err := send(peer, "WELCOME", peer.pseudo); err != nil {
		fmt.Println("Error writing socket ", err)
		delete(peers, peer.address)
		return
	}

	for {
		message, err := network.GetNextMessage(conn)
		// bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message ", err)
			delete(peers, peer.address)
			for _, p := range peers {
				sendPeers(p)
			}
			return
		}

		fmt.Println(conn.RemoteAddr(), "said :", message)

		switch message.Kind {
		case "HELLO":
			handleHello(peer, message.Data)
		default:
			fmt.Println("Unknown message type", message)
		}
	}
}

func handleHello(peer Peer, data string) {
	port, err := strconv.Atoi(strings.TrimSpace(data))

	if err != nil {
		delete(peers, peer.address)
		fmt.Println("Invalid HELLO message ", err)
		return
	}

	peer.chatPort = port
	peer.chatAddress = strings.SplitN(peer.address, ":", 2)[0] + ":" + strings.TrimSpace(data)
	peers[peer.address] = peer

	for _, p := range peers {
		sendPeers(p)
	}
	for addr, p := range peers {
		if addr != peer.address {
			if _, err := send(peer, "NAME", p.chatAddress+" "+p.pseudo); err != nil {
				fmt.Println("Error writing socket ", err)
				delete(peers, peer.address)
				return
			}
			if _, err := send(p, "NAME", peer.chatAddress+" "+peer.pseudo); err != nil {
				fmt.Println("Error writing socket ", err)
			}
		}
	}
}

func sendPeers(peer Peer) {
	list := ""
	for addr, p := range peers {
		if addr == peer.address {
			continue
		}

		list = list + p.chatAddress + " "
	}

	if _, err := send(peer, "PEERS", list); err != nil {
		fmt.Println("Error writing socket ", err)
		delete(peers, peer.address)
		return
	}
}

func send(peer Peer, msgType string, data string) (int, error) {
	time.Sleep(10 * time.Millisecond)
	return peer.conn.Write([]byte(msgType + " " + data + "\n"))
}
