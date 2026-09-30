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
	body := n.Message
	if n.ActionHint != "" {
		body += " - Action: " + n.ActionHint
	}
	cleanMsg := strings.ReplaceAll(body, `\`, `\\`)
	cleanMsg = strings.ReplaceAll(cleanMsg, `"`, `\"`)
	cleanTitle := strings.ReplaceAll(n.Title, `\`, `\\`)
	cleanTitle = strings.ReplaceAll(cleanTitle, `"`, `\"`)

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, cleanMsg, cleanTitle)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("darwin osascript notification failed: %w", err)
	}
	return nil
}
