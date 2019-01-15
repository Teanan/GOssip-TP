// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package browser

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/toqueteos/webbrowser"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connectedWebpage *Webpage

// Webpage is a middleman between the websocket connection and Go.
type Webpage struct {
	Disconnected chan bool

	receiveCallback func(string)

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages (from Go to Browser)
	send chan string

	// Buffered channel of inbound messages (from Browser to Go)
	receive chan string
}

// receiveLoop pumps messages from the websocket to the Browser.receive channel.
func (wpage *Webpage) receiveLoop() {
	defer func() {
		wpage.conn.Close()
		wpage.Disconnected <- true
	}()
	wpage.conn.SetReadLimit(maxMessageSize)
	wpage.conn.SetReadDeadline(time.Now().Add(pongWait))
	wpage.conn.SetPongHandler(func(string) error { wpage.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := wpage.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		wpage.receive <- string(message)
		if wpage.receiveCallback != nil {
			wpage.receiveCallback(string(message))
		}
	}
}

// sendLoop pumps messages from the Browser.send channel to the websocket.
func (wpage *Webpage) sendLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		wpage.conn.Close()
		wpage.Disconnected <- true
	}()
	for {
		select {
		case message, ok := <-wpage.send:
			byteMessage := []byte(message)
			wpage.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				wpage.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := wpage.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			writer.Write(byteMessage)

			// Add queued chat messages to the current websocket message.
			n := len(wpage.send)
			for i := 0; i < n; i++ {
				writer.Write(newline)
				writer.Write([]byte(<-wpage.send))
			}

			if err := writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			wpage.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wpage.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a message to the browser
func (wpage *Webpage) SendMessage(message string) {
	// TODO : explore b.conn.WriteJSON;
	wpage.send <- message
}

// OnReceiveMessage registers a callback for messages received from the browser
func (wpage *Webpage) OnReceiveMessage(callback func(string)) {
	wpage.receiveCallback = callback
}

// ReceiveChannel returns a channel used to read data from the browser
func (wpage *Webpage) ReceiveChannel() <-chan string {
	return wpage.receive
}

// SendChannel returns a channel used to send data to the browser
func (wpage *Webpage) SendChannel() chan<- string {
	return wpage.send
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if connectedWebpage != nil {
		http.Error(w, "Browser already opened and connected", http.StatusTeapot)
		return
	}
	http.ServeFile(w, r, "home.html")
}

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request, connected chan *Webpage) {

	if connectedWebpage != nil {
		http.Error(w, "Browser already opened and connected", http.StatusTeapot)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		connected <- nil
		return
	}

	wpage := &Webpage{conn: conn, send: make(chan string, 256), receive: make(chan string, 256), Disconnected: make(chan bool)}
	go wpage.sendLoop()
	go wpage.receiveLoop()

	connected <- wpage
}

// Connect : connect to browser
func Connect(host string, port int) (*Webpage, error) {

	addr := host + ":" + strconv.Itoa(port)

	if connectedWebpage != nil {
		return nil, errors.New("Already connected to browser")
	}

	connected := make(chan *Webpage)
	go startServer(addr, connected)

	connectedWebpage = <-connected
	if connectedWebpage == nil {
		return nil, errors.New("Couldn't connect to browser")
	}

	return connectedWebpage, nil
}

func startServer(addr string, connected chan *Webpage) {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r, connected)
	})

	go func() {
		time.Sleep(200 * time.Millisecond)
		fullAddr := "http://" + addr + "/"
		webbrowser.Open(fullAddr)
		fmt.Println("HTTP server listening on " + fullAddr)
	}()

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		connected <- nil
	}
}
