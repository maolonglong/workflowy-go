package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func (a *app) newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	login := &cobra.Command{
		Use:   "login",
		Short: "Save API key to config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var apiKey string
			if term.IsTerminal(int(os.Stdin.Fd())) {
				fmt.Fprint(os.Stderr, "Enter API key: ")
				raw, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Fprintln(os.Stderr)
				if err != nil {
					return fmt.Errorf("failed to read API key: %w", err)
				}
				apiKey = string(raw)
			} else {
				buf, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read API key from stdin: %w", err)
				}
				apiKey = strings.TrimSpace(string(buf))
			}
			if apiKey == "" {
				return fmt.Errorf("API key cannot be empty")
			}
			if err := saveConfig(&Config{APIKey: apiKey}); err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, map[string]string{"status": "ok"})
			} else {
				dir, _ := configDir()
				fmt.Printf("API key saved to %s\n", dir)
			}
			return nil
		},
	}

	cmd.AddCommand(login)
	return cmd
}
