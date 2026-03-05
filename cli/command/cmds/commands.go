package cmds

import (
	"fmt"

	"github.com/jgfranco17/lazyfile/cli/internal/errorhandling"
	"github.com/jgfranco17/lazyfile/cli/internal/logging"
	"github.com/spf13/cobra"
)

type runCommand struct{}

func (c *runCommand) invoke() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the application",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx)
			logger.Info("Running the application")
			return &errorhandling.CommandError{
				Err:      fmt.Errorf("not implemented"),
				ExitCode: 1,
				HelpText: "This command is not implemented yet. Please check back later.",
			}
		},
	}
}

func NewRunCommand() *cobra.Command {
	cmd := &runCommand{}
	return cmd.invoke()
}
