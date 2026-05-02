package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

type app struct {
	jsonOutput bool
	maxOutput  int
	client     *workflowy.Client
}

func (a *app) initClient() error {
	apiKey := os.Getenv("WF_API_KEY")
	if apiKey == "" {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		apiKey = cfg.APIKey
	}
	if apiKey == "" {
		return fmt.Errorf("not authenticated. Set WF_API_KEY or run: wf auth login <api-key>")
	}
	client, err := workflowy.NewClient(workflowy.WithAPIKey(apiKey))
	if err != nil {
		return err
	}
	a.client = client
	return nil
}

// Execute runs the root command and returns the exit code.
func Execute() int {
	a := &app{}

	root := &cobra.Command{
		Use:           "wf",
		Short:         "AI-friendly CLI for WorkFlowy",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return a.initClient()
		},
	}

	root.PersistentFlags().BoolVar(&a.jsonOutput, "json", false, "output in JSON format")
	root.PersistentFlags().IntVar(&a.maxOutput, "max-output", 0, "max output characters (0 = unlimited)")

	auth := a.newAuthCmd()
	auth.PersistentPreRunE = func(cmd *cobra.Command, args []string) error { return nil }

	root.AddCommand(
		auth,
		a.newCreateCmd(),
		a.newGetCmd(),
		a.newListCmd(),
		a.newUpdateCmd(),
		a.newDeleteCmd(),
		a.newCompleteCmd(),
		a.newUncompleteCmd(),
		a.newMoveCmd(),
		a.newIDCmd(),
		a.newTreeCmd(),
		a.newSearchCmd(),
		a.newTargetsCmd(),
	)

	if err := root.Execute(); err != nil {
		if a.jsonOutput {
			printError(os.Stderr, err, "command_error")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return 1
	}
	return 0
}
