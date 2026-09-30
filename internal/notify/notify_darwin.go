//go:build darwin

package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

type darwinNotifier struct{}

func NewNotifier() Notifier {
	return &darwinNotifier{}
}

func (d *darwinNotifier) Send(n Notification) error {
	cleanMsg := strings.ReplaceAll(n.Message, `"`, `\"`)
	cleanTitle := strings.ReplaceAll(n.Title, `"`, `\"`)

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, cleanMsg, cleanTitle)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("darwin osascript notification failed: %w", err)
	}
	return nil
}
