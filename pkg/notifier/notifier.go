/*
The notifier package

	This is the interface a plugin has to implement to be usable by this
	project as notification provider. The plugin has to expose an already
	initialized symbol called Provider that implements the Notifier interface
	here defined.
*/
package notifier

import (
	"context"
	"encoding/json"
)

type Notifier interface {
	// Init has to be invoked before the Notifier becomes available to send notifications.
	Init(ctx context.Context) error
	// Notify is the method to invoke in order to send notifications. It might either
	// be blocking, for reliable notifications, or non blocking, for fire-and-forget
	// notifications.
	Notify(ctx context.Context, event NotificationEvent, metadata map[string]string) error
}

type NotificationEvent int

const (
	NotifyAccountCreated NotificationEvent = iota
	NotifyAccountUpdated
	NotifyAccountRemoved
)

func (n NotificationEvent) String() string {
	switch n {
	case NotifyAccountCreated:
		return "CREATED"
	case NotifyAccountUpdated:
		return "UPDATED"
	case NotifyAccountRemoved:
		return "REMOVED"
	default:
		return "UNKNOWN"
	}
}

func (n NotificationEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}
