package cli

import (
	"log"

	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/credentials"
	"github.com/spf13/cobra"
)

var (
	loginUser string
	loginPass string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to FortiOS Captive Portal",
	Long:  "Authenticate to the FortiOS Captive Portal and store the credentials",
	Run:   runLogin,
}

func runLogin(cmd *cobra.Command, args []string) {

	var user, pass string
	var err error

	if loginUser != "" {
		user = loginUser
		if loginPass != "" {
			pass = loginPass
		} else {
			pass, err = credentials.PromptPassword(user)
		}
	} else {
		user, pass, err = credentials.PromptCredentials()
		if loginPass != "" {
			pass = loginPass
		}
	}

	if err != nil {
		log.Fatalf("[ERROR] Input Error: %v", err)
	}

	if user == "" || pass == "" {
		log.Fatalf("[ERROR] Input Error: username and password cannot be empty")
	}

	if err := credentials.Set(user, pass); err != nil {
		log.Fatalf("[FATAL] Failed to save credentials to OS Keyring: %v", err)
	}
	log.Printf("[SUCCESS] Encrypted credentials for %s saved to system keyring.", user)

	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Could not load existing config, creating fresh: %v", err)
		cfg = config.Default()
	}

	cfg.Username = user
	if err := config.Save(cfg); err != nil {
		log.Printf("[WARN] Failed to save default user to config: %v", err)
	} else {
		log.Printf("[INFO] Updated active user to %s in config.", user)
	}
}

func init() {
	loginCmd.Flags().StringVarP(&loginUser, "user", "u", "", "Set or override kawaii-wify username")
	loginCmd.Flags().StringVarP(&loginPass, "password", "p", "", "Set kawaii-wify password directly (non-interactive)")
	rootCmd.AddCommand(loginCmd)
}
