package credentials

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/zalando/go-keyring"
)

func Resolve(explicitUser string) (string, string, error) {
	envUser := os.Getenv("KAWAII_USER")
	envPass := os.Getenv("KAWAII_PASS")

	// 1. Determine target user: Flag > Env
	targetUser := explicitUser
	if targetUser == "" {
		targetUser = envUser
	}

	// 2. If we have a password from ENV and it matches our target user (or user was set by ENV)
	if envPass != "" && (explicitUser == "" || explicitUser == envUser) && targetUser != "" {
		return targetUser, envPass, nil
	}

	// 3. Look up password in Keyring
	if targetUser != "" {
		pass, err := Get(targetUser)
		if err == nil {
			return targetUser, pass, nil
		}
		if !errors.Is(err, keyring.ErrNotFound) {
			return "", "", fmt.Errorf("[ERROR] failed to access system keyring: %w", err)
		}
	}

	// 4. Fallback to interactive prompt
	var (
		user string
		pass string
		err  error
	)
	if targetUser != "" {
		user = targetUser
		pass, err = PromptPassword(user)
		if err != nil {
			return "", "", err
		}
	} else {
		user, pass, err = PromptCredentials()
		if err != nil {
			return "", "", err
		}
	}

	err = Set(user, pass)
	if err != nil {
		log.Printf("[WARN] Failed to save credentials to OS keyring: %v", err)
	}
	return user, pass, nil
}
