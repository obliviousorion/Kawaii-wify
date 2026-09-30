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
	log.Printf("[NOTIFY] [%s] %s: %s", n.Category, n.Title, n.Message)
	return nil
}
