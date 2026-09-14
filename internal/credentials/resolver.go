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

	// If both are present, we have everything we need. Return immediately.
	if envUser != "" && envPass != "" {
		return envUser, envPass, nil
	}

	targetUser := explicitUser
	
	if targetUser == "" {
    	targetUser = envUser
	}

	if targetUser != "" {
		pass, err := Get(targetUser)
		if err == nil {
			return targetUser, pass, nil
		}

		// If D-Bus or the system keyring crashed, stop and report the error
		if !errors.Is(err, keyring.ErrNotFound) {
			return "", "", fmt.Errorf("[ERROR] failed to access system keyring: %w", err)
		}

		// If it was ErrNotFound, don't return an error!
		// Simply let execution fall through to the prompt below.
	}

	user, pass, err := PromptCredentials()
	if err != nil {
		return "", "", err
	}
	err = Set(user, pass)
	if err != nil {
		log.Printf("[ERROR] error while saving credentials to keyring: %v", err)
	}
	return user, pass, nil	
}