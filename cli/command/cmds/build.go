package cmds

import (
	"fmt"
	"io/fs"
	"time"

	"github.com/fatih/color"
	"github.com/jgfranco17/hackstack/cli/internal/errorhandling"
	"github.com/jgfranco17/hackstack/cli/internal/fileutils"
	"github.com/jgfranco17/hackstack/cli/internal/logging"
	"github.com/jgfranco17/hackstack/cli/internal/templating"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type buildCommand struct {
	engineFactory templaterFactory
	outputPath    string
}

func NewBuildCommand() *cobra.Command {
	cmd := &buildCommand{
		engineFactory: templating.NewEngine,
	}
	return cmd.invoke()
}

func (c *buildCommand) invoke() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a new project from a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx).WithField("command", "build")
			category := args[0]

			files, data, err := templating.Load(ctx, category)
			if err != nil {
				return &errorhandling.CommandError{
					Err:      fmt.Errorf("failed to load template resources for category %q: %w", category, err),
					ExitCode: errorhandling.ExitGenericError,
					HelpText: "Failed to load template resources. Please check the category name and try again.",
				}
			}
			if !fileutils.IsDir(c.outputPath) {
				return &errorhandling.CommandError{
					Err:      fmt.Errorf("output path %q is not a directory", c.outputPath),
					ExitCode: errorhandling.ExitInputError,
					HelpText: "The specified output path is not a directory, please provide a valid directory path.",
				}
			}

			logger.WithFields(logrus.Fields{
				"category": category,
				"output":   c.outputPath,
			}).Trace("Beginning template render")
			startTime := time.Now()
			templater := c.engineFactory(files, data)
			if err := templater.Render(ctx, c.outputPath); err != nil {
				return &errorhandling.CommandError{
					Err:      fmt.Errorf("failed to render template %q: %w", category, err),
					ExitCode: errorhandling.ExitGenericError,
					HelpText: "Please check the template and try again.",
				}
			}
			logger.WithFields(logrus.Fields{
				"category": category,
				"output":   c.outputPath,
				"duration": time.Since(startTime).String(),
			}).Info("Template render completed")

			greenFmt := color.New(color.FgGreen).FprintlnFunc()
			greenFmt(cmd.OutOrStdout(), "Built new templated project, happy hacking!")

			return nil
		},
	}

	cmd.Flags().StringVarP(&c.outputPath, "output", "o", ".", "Output directory for the generated project")
	return cmd
}

type templaterFactory func(files fs.FS, data templating.DynamicData) *templating.Engine
