package cli

import (
	"fmt"
	"os"

	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/obliviousorion/kawaii-wify/internal/credentials"
	"github.com/obliviousorion/kawaii-wify/internal/ipc"
	"github.com/spf13/cobra"
)

var (
	loginUser string
	loginPass string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to FortiOS Captive Portal",
	Long:  "Store credentials in OS keyring and authenticate to the FortiOS captive portal.",
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if user == "" || pass == "" {
		fmt.Fprintln(os.Stderr, "Error: username and password cannot be empty")
		os.Exit(1)
	}

	if err := credentials.Set(user, pass); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to save credentials to OS keyring: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	cfg.Username = user
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save default user to config: %v\n", err)
	}

	fmt.Printf("✓ Credentials for %s saved to system keyring.\n", user)

	// If background daemon is running, command it to connect immediately
	client, err := ipc.NewClient()
	if err == nil {
		defer client.Close()
		fmt.Println("Connecting to network via background daemon...")
		resp, err := client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: connect request failed: %v\n", err)
		} else if !resp.Success {
			fmt.Fprintf(os.Stderr, "✕ %s\n", resp.Message)
		} else {
			fmt.Printf("✓ %s\n", resp.Message)
		}
	} else {
		fmt.Printf("✓ Active user set to %s. Run 'kawaii-wify daemon' or 'kawaii-wify connect' to authenticate.\n", user)
	}
}

func init() {
	loginCmd.Flags().StringVarP(&loginUser, "user", "u", "", "Set or override kawaii-wify username")
	loginCmd.Flags().StringVarP(&loginPass, "password", "p", "", "Set kawaii-wify password directly (non-interactive)")
	rootCmd.AddCommand(loginCmd)
}
