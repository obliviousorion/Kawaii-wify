package cli

import (
	"errors"
	"log"

	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/credentials"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var logoutUser string

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and purge stored credentials",
	Long:  "Purges stored credentials from the OS keyring and clears the active identity from configuration.",
	Run:   runLogout,
}

func runLogout(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[WARN] Failed to load config, using defaults: %v", err)
		cfg = config.Default()
	}

	targetUser := logoutUser
	if targetUser == "" {
		targetUser = cfg.Username
	}

	if targetUser == "" {
		log.Println("[INFO] No active user session found.")
		return
	}

	if err := credentials.Delete(targetUser); err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			log.Fatalf("[FATAL] Failed to remove credentials from keyring: %v", err)
		}
	}

	if cfg.Username == targetUser {
		cfg.Username = ""
		if err := config.Save(cfg); err != nil {
			log.Printf("[WARN] Failed to clear user from config: %v", err)
		}
	}

	log.Printf("[SUCCESS] Logged out and removed credentials for %s.", targetUser)
}

func init() {
	logoutCmd.Flags().StringVarP(&logoutUser, "user", "u", "", "Set or override kawaii-wify username to log out")
	rootCmd.AddCommand(logoutCmd)
}
