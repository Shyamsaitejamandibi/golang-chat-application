package main

import (
	"bytes"
	"log"
	"sync"
	"text/template"
)

type Message struct {
	ClientID string `json:"clientID"`
	Text     string `json:"body"`
}
type WSMessage struct {
	Text    string      `json:"text"`
	headers interface{} `json:"headers"`
}

type Hub struct {
	sync.RWMutex
	clients    map[*Client]bool
	messages   []*Message
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client %s connected", client.id)

			for _, message := range h.messages {
				client.send <- getMessageTemplate(message)
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			h.messages = append(h.messages, message)

			for client := range h.clients {
				select {
				case client.send <- getMessageTemplate(message):
				default:

					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func getMessageTemplate(message *Message) []byte {
	tmpl, err := template.ParseFiles("templates/message.html")
	if err != nil {
		log.Println(err)
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, message)
	if err != nil {
		log.Println(err)
	}
	return tpl.Bytes()
}
