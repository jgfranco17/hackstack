package cmds

import (
	"fmt"

	"github.com/jgfranco17/hackstack/cli/internal/errorhandling"
	"github.com/jgfranco17/hackstack/cli/internal/logging"
	"github.com/spf13/cobra"
)

type buildCommand struct{}

func (c *buildCommand) invoke() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build a new project from a template",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx)
			logger.Info("Building a new project from a template")
			return &errorhandling.CommandError{
				Err:      fmt.Errorf("not implemented"),
				ExitCode: 1,
				HelpText: "This command is not implemented yet. Please check back later.",
			}
		},
	}
}

func NewBuildCommand() *cobra.Command {
	cmd := &buildCommand{}
	return cmd.invoke()
}
