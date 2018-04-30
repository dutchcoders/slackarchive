// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 2 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 16384
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	api *api

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *connection) Close() {
	defer func() {
		c.api.unregister <- c
		c.ws.Close()

		log.Info("Connection closed")
	}()
}

type Message struct {
	Category string
	Body     json.RawMessage
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	// the client wants to know if the server is there, the server doesn't need to know the client isn't there.?
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.ws.ReadMessage()
		if err == nil {
		} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			log.Error("Connection closed unexpectedly: %s", err.Error())
			break
		} else {
			log.Error("Connection error: %s", err.Error())
			break
		}

		var msg Message
		if err = json.Unmarshal(message, &msg); err != nil {
			log.Error("error: %v", err)
			continue
		}

		c.api.indexChan <- msg
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.write(websocket.TextMessage, message); err != nil {
				log.Errorf("writePump error: %s", err.Error())
				return
			}

			log.Infof("%#v", message)

		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				log.Error("writePump error: ", err.Error())
				return
			}
		}
	}
}
