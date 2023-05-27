package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/icrowley/fake"
	uuid "github.com/satori/go.uuid"
)

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	id     string
	name   string
	socket *websocket.Conn
	send   chan []byte
}

type Message struct {
	Sender    string `json:"sender,omitempty"`
	Name      string `json:"name,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var manager = ClientManager{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			jsonMsg, _ := json.Marshal(&Message{Content: "/A socket has connected."})
			manager.send(jsonMsg, conn)

		case conn := <-manager.unregister:

			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMsg, _ := json.Marshal("/A socket has been disconnected.")
				manager.send(jsonMsg, conn)
			}
		case message := <-manager.broadcast:
			// manager.send(message, nil)
			// Change the logic
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

func (manager *ClientManager) send(msg []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- msg
		}
	}
}

func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}

		jsonMsg, _ := json.Marshal(&Message{Sender: c.id, Name: c.name, Content: string(msg)})
		manager.broadcast <- jsonMsg // Check this
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.socket.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func main() {
	fmt.Println("Starting Application...")
	go manager.start()

	http.HandleFunc("/ws", wsPage)
	http.ListenAndServe(":8080", nil)
}

func wsPage(res http.ResponseWriter, req *http.Request) {

	conn, _ := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)

	client := &Client{id: uuid.NewV4().String(), name: fake.FullName(), socket: conn, send: make(chan []byte)}
	manager.register <- client

	go client.read()
	go client.write()
}
