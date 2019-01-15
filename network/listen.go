package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type MessageReceiver interface {
	Receive(Message, Peer)
	HandleHello(data string, from Peer)
}

// Listen ...
func Listen(port int, peers PeersMap, messageReceiver MessageReceiver) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Failed to open listen socket", err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection peer", err)
			return
		}
		go handleConnection(conn, peers, messageReceiver)
	}
}

func handleConnection(conn net.Conn, peers PeersMap, messageReceiver MessageReceiver) {
	var remotePeerAddress = ""
	for {
		message, err := GetNextMessage(conn)
		if err != nil {
			fmt.Println("Failed to read message from peer", err)
			return
		}
		fmt.Println("Got :", message)

		if message.Kind == "HELLO" {
			handleHello(message.Data, conn, peers, &remotePeerAddress, messageReceiver, 5)
		} else {
			handleMessage(&remotePeerAddress, message, peers, messageReceiver, 5)
		}
	}
}

func handleMessage(remotePeerAddress *string, message Message, peers PeersMap, messageReceiver MessageReceiver, retries int) {
	if ok, _ := peers.Find(*remotePeerAddress); !ok {
		if retries == 0 {
			fmt.Println("Got message", message, "from unknown peer", *remotePeerAddress)
		} else {
			go func() {
				time.Sleep(1 * time.Second)
				handleMessage(remotePeerAddress, message, peers, messageReceiver, retries-1)
			}()
		}
	} else {
		messageReceiver.Receive(message, peers.Get(*remotePeerAddress))
	}
}

func handleHello(data string, conn net.Conn, peers PeersMap, remotePeerAddress *string, messageReceiver MessageReceiver, retries int) {
	port, err := strconv.Atoi(strings.TrimSpace(data))

	if err != nil {
		fmt.Println("Invalid HELLO message ", err)
		return
	}

	addr := strings.Split(conn.RemoteAddr().String(), ":")[0] + ":" + strconv.Itoa(port)

	found, p := peers.Find(addr)

	if !found {
		if retries == 0 {
			fmt.Println("Unknown peer", addr)
		} else {
			go func() {
				time.Sleep(1 * time.Second)
				handleHello(data, conn, peers, remotePeerAddress, messageReceiver, retries-1)
			}()
		}
		return
	}

	*remotePeerAddress = p.FullAddress()

	fmt.Println("Identified", conn.RemoteAddr(), "as", p)

	messageReceiver.HandleHello(data, peers.Get(*remotePeerAddress))
}
