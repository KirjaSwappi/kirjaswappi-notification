package domain

import "time"

type Notification struct {
	UserID  string
	Title   string
	Message string
	Time    time.Time
}
