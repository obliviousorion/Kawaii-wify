package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/obliviousorion/kawaii-wify/internal/config"
	"github.com/spf13/cobra"
)

var (
	logLines int
	logPath  bool
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View recent daemon log output",
	Long:  "Displays recent log output from the background daemon session, or prints the path to the log file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, err := config.LogFilePath()
		if err != nil {
			return fmt.Errorf("could not determine log file path: %w", err)
		}

		if logPath {
			fmt.Println(filePath)
			return nil
		}

		file, err := os.Open(filePath)
		if os.IsNotExist(err) {
			fmt.Println("No log file found. The daemon has not run yet.")
			fmt.Printf("Expected location: %s\n", filePath)
			return nil
		} else if err != nil {
			return fmt.Errorf("could not open log file: %w", err)
		}
		defer file.Close()

		if logLines <= 0 {
			logLines = 30
		}

		var lines []string
		scanner := bufio.NewScanner(file)
		// Set buffer to 512KB to safely scan lines containing verbose portal responses
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 512*1024)

		for scanner.Scan() {
			lines = append(lines, scanner.Text())
			if len(lines) > logLines {
				lines = lines[1:]
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading log file: %w", err)
		}

		if len(lines) == 0 {
			fmt.Println("(log file is empty)")
			return nil
		}

		for _, l := range lines {
			fmt.Println(l)
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().IntVarP(&logLines, "lines", "n", 30, "Number of recent lines to display")
	logsCmd.Flags().BoolVarP(&logPath, "path", "p", false, "Print the path to the log file and exit")
	rootCmd.AddCommand(logsCmd)
}
