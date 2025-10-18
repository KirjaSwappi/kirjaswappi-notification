package service

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/kirjaswappi/kirjaswappi-notification/internal/domain"
)

func TestBroadcaster(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broadcaster := NewBroadcaster(logger)
	defer broadcaster.Close()

	userID := "test-user"

	// Subscribe
	ch := broadcaster.Subscribe(userID)

	// Send notification
	notification := domain.Notification{
		UserID:  userID,
		Title:   "Test",
		Message: "Test message",
		Time:    time.Now(),
	}

	broadcaster.Broadcast(notification)

	// Receive notification
	select {
	case received := <-ch:
		if received.UserID != userID {
			t.Errorf("Expected userID %s, got %s", userID, received.UserID)
		}
		if received.Title != "Test" {
			t.Errorf("Expected title 'Test', got %s", received.Title)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for notification")
	}

	// Test stats
	users, subs := broadcaster.Stats()
	if users != 1 {
		t.Errorf("Expected 1 user, got %d", users)
	}
	if subs != 1 {
		t.Errorf("Expected 1 subscriber, got %d", subs)
	}

	// Unsubscribe
	broadcaster.Unsubscribe(userID, ch)

	// Check stats after unsubscribe
	users, subs = broadcaster.Stats()
	if users != 0 {
		t.Errorf("Expected 0 users after unsubscribe, got %d", users)
	}
	if subs != 0 {
		t.Errorf("Expected 0 subscribers after unsubscribe, got %d", subs)
	}
}

func TestBroadcasterMultipleSubscribers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broadcaster := NewBroadcaster(logger)
	defer broadcaster.Close()

	userID := "test-user"

	// Subscribe multiple times
	ch1 := broadcaster.Subscribe(userID)
	ch2 := broadcaster.Subscribe(userID)

	// Send notification
	notification := domain.Notification{
		UserID:  userID,
		Title:   "Test",
		Message: "Test message",
		Time:    time.Now(),
	}

	broadcaster.Broadcast(notification)

	// Both should receive
	received1 := <-ch1
	received2 := <-ch2

	if received1.Title != "Test" || received2.Title != "Test" {
		t.Error("Both subscribers should receive the notification")
	}

	// Test stats
	users, subs := broadcaster.Stats()
	if users != 1 {
		t.Errorf("Expected 1 user, got %d", users)
	}
	if subs != 2 {
		t.Errorf("Expected 2 subscribers, got %d", subs)
	}
}
