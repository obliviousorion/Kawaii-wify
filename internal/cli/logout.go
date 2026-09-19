package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/credentials"
	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var logoutUser string

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and purge stored credentials",
	Long:  "Purges stored credentials from the OS keyring, disconnects any active daemon session, and clears the active identity from configuration.",
	Run:   runLogout,
}

func runLogout(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config, using defaults: %v\n", err)
		cfg = config.Default()
	}

	targetUser := logoutUser
	if targetUser == "" {
		targetUser = cfg.Username
	}

	if targetUser == "" {
		fmt.Println("No active user session found.")
		return
	}

	// Notify background daemon if running
	client, err := ipc.NewClient()
	if err == nil {
		defer client.Close()
		resp, err := client.Disconnect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to notify background daemon: %v\n", err)
		} else if !resp.Success {
			fmt.Fprintf(os.Stderr, "Warning: daemon disconnect rejected: %s\n", resp.Message)
		} else {
			fmt.Println("✓ Background daemon disconnected and session revoked.")
		}
	}

	if err := credentials.Delete(targetUser); err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "Error: failed to remove credentials from keyring: %v\n", err)
			os.Exit(1)
		}
	}

	if cfg.Username == targetUser {
		cfg.Username = ""
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clear user from config: %v\n", err)
		}
	}

	fmt.Printf("✓ Logged out and removed credentials for %s.\n", targetUser)
}

func init() {
	logoutCmd.Flags().StringVarP(&logoutUser, "user", "u", "", "Set or override kawaii-wify username to log out")
	rootCmd.AddCommand(logoutCmd)
}
