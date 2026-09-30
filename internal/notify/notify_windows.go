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
			var uriBuf strings.Builder
			_ = xml.EscapeText(&uriBuf, []byte(uri))
			imageBinding = fmt.Sprintf(`<image placement="appLogoOverride" hint-crop="circle" src="%s"/>`, uriBuf.String())
		}
	}

	xmlContent := fmt.Sprintf(`<toast duration="short"><visual><binding template="ToastGeneric">%s<text>%s</text><text>%s</text></binding></visual><audio src="ms-winsoundevent:Notification.Default"/></toast>`,
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

$appId = 'Kawaii-Wify'
$regPath = 'HKCU:\Software\Classes\AppUserModelId\' + $appId
if (-not (Test-Path $regPath)) {
    try {
        New-Item -Path $regPath -Force | Out-Null
        Set-ItemProperty -Path $regPath -Name 'DisplayName' -Value 'Kawaii-Wify'
    } catch {}
}

$notifier = $null
try {
    $n = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($appId)
    if ($n.Setting -eq [Windows.UI.Notifications.NotificationSetting]::Enabled) {
        $notifier = $n
    }
} catch {}

if ($null -eq $notifier) {
    $fallbackId = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
    $notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($fallbackId)
}

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
