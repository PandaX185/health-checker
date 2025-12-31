package monitor

import "time"

type Event interface {
	Type() string
	At() time.Time
}

type StatusChangeEvent struct {
	ServiceID int
	OldStatus string
	NewStatus string
	Timestamp time.Time
}

func (e StatusChangeEvent) Type() string {
	return "StatusChange"
}

func (e StatusChangeEvent) At() time.Time {
	return e.Timestamp
}
