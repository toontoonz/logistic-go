package handler

import (
	"context"
	"time"

	json "github.com/goccy/go-json"
	"github.com/gofiber/contrib/websocket"
	"github.com/toontoonz/logistic-go/backend/internal/model"
	"github.com/toontoonz/logistic-go/backend/internal/ws"
)

func (a API) positionsWS(conn *websocket.Conn) {
	client := ws.NewClient()
	a.Hub.Register(client)
	defer a.Hub.Unregister(client)

	// WriteLoop runs beside this function. This goroutine only reads,
	// so we notice when the browser goes away. One writer, one reader.
	go func() {
		client.WriteLoop(func(msg []byte) error {
			return conn.WriteMessage(websocket.TextMessage, msg)
		})
		_ = conn.Close()
	}()

	a.sendSnapshot(client)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (a API) sendSnapshot(client *ws.Client) {
	if a.Locations == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	positions, err := a.Locations.Snapshot(ctx)
	if err != nil {
		return
	}
	if positions == nil {
		positions = []model.Position{}
	}
	body, err := json.Marshal(ws.SnapshotMessage{Type: "snapshot", Positions: positions})
	if err != nil {
		return
	}
	client.TrySend(body)
}
