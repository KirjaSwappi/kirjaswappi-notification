package service

import (
	"log/slog"
	"sync"

	"github.com/kirjaswappi/kirjaswappi-notification/internal/domain"
)

type Subscriber chan domain.Notification

type Broadcaster struct {
	subscribers map[string][]Subscriber // userID -> channels
	lock        sync.RWMutex
	logger      *slog.Logger
	closed      bool
}

func NewBroadcaster(logger *slog.Logger) *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[string][]Subscriber),
		logger:      logger,
		closed:      false,
	}
}

func (b *Broadcaster) Subscribe(userID string) Subscriber {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.closed {
		ch := make(Subscriber)
		close(ch)
		return ch
	}

	ch := make(Subscriber, 10)
	b.subscribers[userID] = append(b.subscribers[userID], ch)

	b.logger.Debug("User subscribed",
		slog.String("user_id", userID),
		slog.Int("total_subscribers", len(b.subscribers[userID])))

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

			// Clean up empty user entries
			if len(b.subscribers[userID]) == 0 {
				delete(b.subscribers, userID)
			}

			b.logger.Debug("User unsubscribed",
				slog.String("user_id", userID),
				slog.Int("remaining_subscribers", len(b.subscribers[userID])))
			break
		}
	}
}

func (b *Broadcaster) Broadcast(n domain.Notification) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.closed {
		return
	}

	subscribers := b.subscribers[n.UserID]
	if len(subscribers) == 0 {
		b.logger.Debug("No subscribers for notification",
			slog.String("user_id", n.UserID),
			slog.String("title", n.Title))
		return
	}

	delivered := 0
	dropped := 0

	for _, ch := range subscribers {
		select {
		case ch <- n:
			delivered++
		default:
			// Drop message if channel is full
			dropped++
		}
	}

	b.logger.Debug("Notification broadcast",
		slog.String("user_id", n.UserID),
		slog.String("title", n.Title),
		slog.Int("delivered", delivered),
		slog.Int("dropped", dropped))
}

func (b *Broadcaster) Close() {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.closed {
		return
	}

	b.closed = true

	// Close all subscriber channels
	for userID, subs := range b.subscribers {
		for _, ch := range subs {
			close(ch)
		}
		delete(b.subscribers, userID)
	}

	b.logger.Info("Broadcaster closed")
}

func (b *Broadcaster) Stats() (int, int) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	users := len(b.subscribers)
	totalSubs := 0
	for _, subs := range b.subscribers {
		totalSubs += len(subs)
	}

	return users, totalSubs
}
