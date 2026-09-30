//go:build windows

package notify

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/obliviousorion/kawaii-wify/internal/assets"
)

type windowsNotifier struct{}

func NewNotifier() Notifier {
	return &windowsNotifier{}
}

func (w *windowsNotifier) Send(n Notification) error {
	mascotName := MascotFileForCategory(n.Category)
	mascotPath, err := assets.GetMascotPath(mascotName)
	if err != nil {
		mascotPath = ""
	}

	var titleBuf, msgBuf strings.Builder
	_ = xml.EscapeText(&titleBuf, []byte(n.Title))
	_ = xml.EscapeText(&msgBuf, []byte(n.Message))

	escapedTitle := titleBuf.String()
	escapedMsg := msgBuf.String()

	imageBinding := ""
	if mascotPath != "" {
		if absPath, err := filepath.Abs(mascotPath); err == nil {
			uri := "file:///" + filepath.ToSlash(absPath)
			imageBinding = fmt.Sprintf(`<image placement="appLogoOverride" hint-crop="circle" src="%s"/>`, uri)
		}
	}

	xmlContent := fmt.Sprintf(`<toast><visual><binding template="ToastGeneric">%s<text>%s</text><text>%s</text></binding></visual></toast>`,
		imageBinding, escapedTitle, escapedMsg)

	// PowerShell script to invoke Windows WinRT Toast
	psScript := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml(@'
%s
'@)

$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
$appId = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($appId)
$notifier.Show($toast)
`, xmlContent)

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", "-")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdin = strings.NewReader(psScript)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell toast notification failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}
