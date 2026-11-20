// Package ws fans driver positions out to connected browsers.
package ws

import "context"

// Client is one browser socket. send is buffered so a slow browser
// does not stall every other subscriber.
type Client struct {
	send chan []byte
}

// NewClient builds a client with room for a short burst of updates.
func NewClient() *Client {
	return &Client{send: make(chan []byte, 32)}
}

// TrySend queues a message. If the buffer is full or the client is
// already gone, the update is dropped — the next GPS ping will refresh it.
func (c *Client) TrySend(msg []byte) {
	select {
	case c.send <- msg:
	default:
	}
}

// WriteLoop is the only goroutine allowed to write to conn.
// WebSocket connections are not safe for concurrent writers.
func (c *Client) WriteLoop(write func([]byte) error) {
	for msg := range c.send {
		if err := write(msg); err != nil {
			return
		}
	}
}

type registration struct {
	client *Client
	done   chan struct{}
}

// Hub owns the set of connected browsers.
//
// Run is the only goroutine that touches the clients map. Other
// goroutines (HTTP handlers) never take a lock; they send on a channel
// and Run handles the map. That is the usual Go way to share state:
// "do not communicate by sharing memory; share memory by communicating."
type Hub struct {
	clients    map[*Client]struct{}
	register   chan registration
	unregister chan *Client
	broadcast  chan []byte
}

// NewHub builds a hub. Call Run before Register.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan registration),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 64),
	}
}

// Run processes hub events until ctx is cancelled (process shutdown).
// Start it with `go hub.Run(ctx)` so it does not block main.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for c := range h.clients {
				close(c.send)
			}
			return
		case reg := <-h.register:
			h.clients[reg.client] = struct{}{}
			close(reg.done)
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					// A full buffer means the browser is too slow.
					// Drop it instead of blocking the fan-out.
					delete(h.clients, c)
					close(c.send)
				}
			}
		}
	}
}

// Register blocks until Run has added the client. The ack channel is
// how the caller knows the add has actually happened: sending on
// h.register only means Run received the value, and the map write
// happens after that receive.
func (h *Hub) Register(c *Client) {
	done := make(chan struct{})
	h.register <- registration{client: c, done: done}
	<-done
}

// Unregister removes a client. Safe to call twice.
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// Broadcast queues a message for every client. If Run is busy and the
// queue is full, the message is dropped so the GPS handler stays fast.
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
	}
}
