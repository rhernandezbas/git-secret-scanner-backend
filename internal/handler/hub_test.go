package handler

import (
	"testing"
	"time"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
)

func TestHub_Register(t *testing.T) {
	hub := NewHub()
	client := make(chan domain.ProgressEvent, 1)

	hub.Register(client)

	hub.mu.RLock()
	_, ok := hub.clients[client]
	hub.mu.RUnlock()

	if !ok {
		t.Fatal("expected client to be registered")
	}
}

func TestHub_Unregister(t *testing.T) {
	hub := NewHub()
	client := make(chan domain.ProgressEvent, 1)

	hub.Register(client)
	hub.Unregister(client)

	hub.mu.RLock()
	_, ok := hub.clients[client]
	hub.mu.RUnlock()

	if ok {
		t.Fatal("expected client to be unregistered")
	}
}

func TestHub_Broadcast_SendsToAllClients(t *testing.T) {
	hub := NewHub()
	c1 := make(chan domain.ProgressEvent, 1)
	c2 := make(chan domain.ProgressEvent, 1)

	hub.Register(c1)
	hub.Register(c2)

	event := domain.ProgressEvent{Repo: "test-repo", Type: domain.EventRepoStart, Message: "cloning"}
	hub.Broadcast(event)

	for _, ch := range []chan domain.ProgressEvent{c1, c2} {
		select {
		case got := <-ch:
			if got.Repo != event.Repo || got.Type != event.Type {
				t.Errorf("unexpected event: got %+v, want %+v", got, event)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timed out waiting for event")
		}
	}
}

func TestHub_Broadcast_NonBlocking_WhenClientFull(t *testing.T) {
	hub := NewHub()
	// unbuffered channel — will be full immediately
	full := make(chan domain.ProgressEvent)
	hub.Register(full)

	event := domain.ProgressEvent{Repo: "test-repo", Type: domain.EventFileScanned, Message: "scanning"}

	done := make(chan struct{})
	go func() {
		hub.Broadcast(event)
		close(done)
	}()

	select {
	case <-done:
		// good — broadcast returned without blocking
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Broadcast blocked on full client channel")
	}
}

func TestHub_Broadcast_MultipleClients_SameEvent(t *testing.T) {
	hub := NewHub()
	n := 5
	clients := make([]chan domain.ProgressEvent, n)
	for i := range clients {
		clients[i] = make(chan domain.ProgressEvent, 1)
		hub.Register(clients[i])
	}

	event := domain.ProgressEvent{Repo: "multi-repo", Type: domain.EventScanComplete, Message: "done"}
	hub.Broadcast(event)

	for i, ch := range clients {
		select {
		case got := <-ch:
			if got.Repo != event.Repo {
				t.Errorf("client %d: unexpected event repo: got %q, want %q", i, got.Repo, event.Repo)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client %d: timed out waiting for event", i)
		}
	}
}
