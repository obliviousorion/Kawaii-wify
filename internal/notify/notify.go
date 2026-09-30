package notify

import (
	"github.com/obliviousorion/kawaii-wify/internal/assets"
	"github.com/obliviousorion/kawaii-wify/internal/config"
)

type Category string

const (
	CategorySecurity   Category = "security"
	CategoryAuthFailed Category = "auth_failed"
	CategoryUpdate     Category = "update"
	CategoryOnline     Category = "online"
	CategoryOffline    Category = "offline"
)

type Notification struct {
	Title      string
	Message    string
	Category   Category
	ActionHint string // Specific remediation instruction (e.g. "Run 'kawaii-wify login'")
	ActionURL  string // Optional URL opened when the user clicks the toast
}

type Notifier interface {
	Send(n Notification) error
}

// MascotFileForCategory returns the embedded mascot filename for a given notification category.
func MascotFileForCategory(cat Category) string {
	switch cat {
	case CategorySecurity:
		return assets.MascotSecurity
	case CategoryAuthFailed:
		return assets.MascotAuthFailed
	case CategoryUpdate:
		return assets.MascotUpdate
	case CategoryOnline:
		return assets.MascotOnline
	case CategoryOffline:
		return assets.MascotOffline
	default:
		return assets.MascotOnline
	}
}

// Send dispatches an OS toast notification if desktop notifications are enabled in config.
func Send(n Notification) error {
	cfg, err := config.Load()
	if err == nil && !cfg.IsNotificationsEnabled() {
		return nil // User has notifications disabled
	}

	notifier := NewNotifier()
	return notifier.Send(n)
}
