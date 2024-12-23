package models

import (
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	ID     string
	Conn   *websocket.Conn
	SendCh chan *Message
}

func (c *Client) Listen() {
	for message := range c.SendCh {
		// Marshal the message to JSON format before sending
		if err := c.Conn.WriteJSON(message); err != nil {
			log.Printf("error sending message to client %s: %v", c.ID, err)
			break
		}
	}
	// Close the connection when the client is done listening
	err := c.Conn.Close()
	if err != nil {
		return
	}
}
