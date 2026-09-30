//go:build !windows && !linux && !darwin

package notify

import (
	"log"
)

type fallbackNotifier struct{}

func NewNotifier() Notifier {
	return &fallbackNotifier{}
}

func (f *fallbackNotifier) Send(n Notification) error {
	msg := n.Message
	if n.ActionHint != "" {
		msg += " | Action: " + n.ActionHint
	}
	log.Printf("[NOTIFY] [%s] %s: %s", n.Category, n.Title, msg)
	return nil
}
