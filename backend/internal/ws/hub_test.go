package ws

import (
	"context"
	"testing"
	"time"
)

func TestHubBroadcastReachesClient(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run owns the client map. The test talks to it only through channels.
	go hub.Run(ctx)

	client := NewClient()
	hub.Register(client)

	payload := []byte(`{"type":"position","lat":13.7}`)
	hub.Broadcast(payload)

	select {
	case got := <-client.send:
		if string(got) != string(payload) {
			t.Fatalf("payload = %s", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broadcast")
	}
}

func TestHubUnregisterStopsDelivery(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	client := NewClient()
	hub.Register(client)
	hub.Unregister(client)

	// Unregister closes send. A receive on a closed, empty channel
	// returns immediately with ok == false.
	select {
	case _, ok := <-client.send:
		if ok {
			t.Fatal("expected closed send channel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for unregister")
	}

	// A second browser still receives updates after the first left.
	other := NewClient()
	hub.Register(other)
	hub.Broadcast([]byte(`{"type":"position"}`))
	select {
	case msg := <-other.send:
		if string(msg) != `{"type":"position"}` {
			t.Fatalf("payload = %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for remaining client")
	}
}
