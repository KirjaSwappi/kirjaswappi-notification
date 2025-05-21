package service

import (
	"sync"

	"github.com/kirjaswappi/kirjaswappi-notification/internal/domain"
)

type Subscriber chan domain.Notification

type Broadcaster struct {
	subscribers map[string][]Subscriber // userID -> channels
	lock        sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[string][]Subscriber),
	}
}

func (b *Broadcaster) Subscribe(userID string) Subscriber {
	ch := make(Subscriber, 10)
	b.lock.Lock()
	b.subscribers[userID] = append(b.subscribers[userID], ch)
	b.lock.Unlock()
	return ch
}

func (b *Broadcaster) Unsubscribe(userID string, ch Subscriber) {
	b.lock.Lock()
	defer b.lock.Unlock()

	subs := b.subscribers[userID]
	for i, c := range subs {
		if c == ch {
			b.subscribers[userID] = append(subs[:i], subs[i+1:]...)
			close(c)
			break
		}
	}
}

func (b *Broadcaster) Broadcast(n domain.Notification) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	for _, ch := range b.subscribers[n.UserID] {
		select {
		case ch <- n:
		default:
			// Drop message if channel is full
		}
	}
}
