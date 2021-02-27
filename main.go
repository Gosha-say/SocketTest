package main

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

const html = "PCFET0NUWVBFIGh0bWw+CjxodG1sIGxhbmc9InJ1Ij4KPGhlYWQ+CiAgICA8bWV0YSBjaGFyc2V0PSJVVEYtOCIgLz4KICAgIDxtZXRhIG5hbWU9InZpZXdwb3J0IiBjb250ZW50PSJ3aWR0aD1kZXZpY2Utd2lkdGgsIGluaXRpYWwtc2NhbGU9MS4wIiAvPgogICAgPG1ldGEgaHR0cC1lcXVpdj0iWC1VQS1Db21wYXRpYmxlIiBjb250ZW50PSJpZT1lZGdlIiAvPgogICAgPHRpdGxlPkdvIFdlYlNvY2tldCBUdXRvcmlhbDwvdGl0bGU+CjwvaGVhZD4KPGJvZHk+CjxoMj5IZWxsbyBXb3JsZDwvaDI+CjxwIGlkPSJvdXRwdXQiPjwvcD4KPGxhYmVsIGZvcj0iaW5wdXQiPk1lc3NhZ2U6IDwvbGFiZWw+PGlucHV0IGlkPSJpbnB1dCIgdHlwZT0idGV4dCIgLz4KPGJ1dHRvbiBvbmNsaWNrPSJzZW5kKCkiPlNlbmQ8L2J1dHRvbj4KPHNjcmlwdD4KICAgIGxldCBzb2NrZXQgPSBuZXcgV2ViU29ja2V0KCJ3czovLzEyNy4wLjAuMTo4MDgwL3dzIik7CiAgICBsZXQgaW5wdXQgPSBkb2N1bWVudC5nZXRFbGVtZW50QnlJZCgiaW5wdXQiKTsKCiAgICBjb25zb2xlLmxvZygiQXR0ZW1wdGluZyBDb25uZWN0aW9uLi4uIik7CgovKiAgICBzb2NrZXQub25vcGVuID0gKCkgPT4gewogICAgICAgIGNvbnNvbGUubG9nKCJTdWNjZXNzZnVsbHkgQ29ubmVjdGVkIik7CiAgICAgICAgc29ja2V0LnNlbmQoIkhpIEZyb20gdGhlIENsaWVudCEiKQogICAgfTsKCiAgICBzb2NrZXQub25jbG9zZSA9IGV2ZW50ID0+IHsKICAgICAgICBjb25zb2xlLmxvZygiU29ja2V0IENsb3NlZCBDb25uZWN0aW9uOiAiLCBldmVudCk7CiAgICAgICAgc29ja2V0LnNlbmQoIkNsaWVudCBDbG9zZWQhIikKICAgIH07Ki8KCiAgICBzb2NrZXQub25lcnJvciA9IGVycm9yID0+IHsKICAgICAgICBjb25zb2xlLmxvZygiU29ja2V0IEVycm9yOiAiLCBlcnJvcik7CiAgICB9OwoKICAgIHNvY2tldC5vbm1lc3NhZ2UgPSBmdW5jdGlvbihldnQpIHsKICAgICAgICBsZXQgb3V0ID0gZG9jdW1lbnQuZ2V0RWxlbWVudEJ5SWQoJ291dHB1dCcpOwogICAgICAgIGNvbnNvbGUubG9nKGV2dCk7CiAgICAgICAgbGV0IG1zZyA9SlNPTi5wYXJzZShldnQuZGF0YSk7CiAgICAgICAgb3V0LmlubmVySFRNTCArPSBtc2cuYm9keSArICc8YnI+JzsKICAgIH07CgogICAgZnVuY3Rpb24gc2VuZCgpIHsKICAgICAgICBzb2NrZXQuc2VuZChpbnB1dC52YWx1ZSk7CiAgICAgICAgaW5wdXQudmFsdWUgPSAiIjsKICAgIH0KCjwvc2NyaXB0Pgo8L2JvZHk+CjwvaHRtbD4="

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
	}
}

func Reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

func Writer(conn *websocket.Conn) {
	for {
		fmt.Println("Sending")
		messageType, r, err := conn.NextReader()
		if err != nil {
			fmt.Println(err)
			return
		}
		w, err := conn.NextWriter(messageType)
		if err != nil {
			fmt.Println(err)
			return
		}
		if _, err := io.Copy(w, r); err != nil {
			fmt.Println(err)
			return
		}
		if err := w.Close(); err != nil {
			fmt.Println(err)
			return
		}
	}
}

func serveWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func setupRoutes() {
	pool := NewPool()
	go pool.Start()
	//tpl, err := ioutil.ReadFile("./tpl.html")
	//tpl64 := b64.StdEncoding.EncodeToString(tpl)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := b64.StdEncoding.DecodeString(html)
		fmt.Fprintf(w, string(body))
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func main() {
	fmt.Println("Chat App v0.02")
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}

func (pool *Pool) Start() {
	log.Println("Start...")
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client := range pool.Clients {
				fmt.Println(client)
				client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client := range pool.Clients {
				client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			}
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		message := Message{Type: messageType, Body: string(p)}
		c.Pool.Broadcast <- message
		fmt.Printf("Message Received: %+v\n", message)
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conn, nil
}
