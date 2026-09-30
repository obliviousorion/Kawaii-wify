//go:build linux

package notify

import (
	"fmt"
	"os/exec"

	"github.com/obliviousorion/kawaii-wify/internal/assets"
)

type linuxNotifier struct{}

func NewNotifier() Notifier {
	return &linuxNotifier{}
}

func (l *linuxNotifier) Send(n Notification) error {
	mascotName := MascotFileForCategory(n.Category)
	mascotPath, _ := assets.GetMascotPath(mascotName)

	args := []string{"-a", "Kawaii-Wify"}
	if mascotPath != "" {
		args = append(args, "-i", mascotPath)
	}
	body := n.Message
	if n.ActionHint != "" {
		body += "\nAction: " + n.ActionHint
	}
	args = append(args, n.Title, body)

	cmd := exec.Command("notify-send", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("linux notify-send failed: %w", err)
	}
	return nil
}
