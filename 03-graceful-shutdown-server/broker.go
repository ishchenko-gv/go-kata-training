package server

import (
	"context"
	"log/slog"
	"sync"
)

type Broker interface {
	Start(ctx context.Context)
	Subscribe(client chan string)
	Unsubscribe(client chan string)
}

type broker struct {
	logger  *slog.Logger
	mu      sync.Mutex
	clients map[chan string]bool
	msgCh   chan string
}

func NewBroker(msgCh chan string) *broker {
	return &broker{
		logger:  slog.Default(),
		clients: make(map[chan string]bool),
	}
}

func (b *broker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-b.msgCh:
			for c := range b.clients {
				c <- msg
			}
		}
	}
}

func (b *broker) Subscribe(client chan string) {
	b.mu.Lock()
	b.clients[client] = true
	b.mu.Unlock()
}

func (b *broker) Unsubscribe(client chan string) {
	b.mu.Lock()
	delete(b.clients, client)
	b.mu.Unlock()
}
