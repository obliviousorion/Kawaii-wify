package credentials

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func PromptCredentials() (string, string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", "", fmt.Errorf("standard input is not a terminal")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	username = strings.TrimSpace(username)

	fmt.Print("Enter Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Print newline after masked input
	if err != nil {
		return "", "", err
	}
	password := string(passwordBytes)

	if username == "" {
		return "", "", fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return "", "", fmt.Errorf("password cannot be empty")
	}

	// this never returns empty strings
	return username, password, nil
}
