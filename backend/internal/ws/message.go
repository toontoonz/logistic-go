package ws

import "github.com/toontoonz/logistic-go/backend/internal/model"

// PositionMessage is one live update pushed to every map.
type PositionMessage struct {
	Type string `json:"type"`
	model.Position
}

// SnapshotMessage is the burst sent when a browser first connects,
// so the map is not blank until the next GPS ping.
type SnapshotMessage struct {
	Type      string           `json:"type"`
	Positions []model.Position `json:"positions"`
}
