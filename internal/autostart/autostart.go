package autostart

// Manager defines the interface for managing system background autostart on boot/login.
type Manager interface {
	Enable() error
	Disable() error
	IsEnabled() (bool, error)
}

// New returns a platform-appropriate autostart manager.
func New() (Manager, error) {
	return newPlatformManager()
}
