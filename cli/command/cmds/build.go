package cmds

import (
	"context"
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
	// Injectable dependencies for easier testing
	engineFactory templaterFactory
	loader        loaderFunc
	dataSource    dataSourceFunc

	// Flags
	outputPath     string
	sourceFile     string
	forceOverwrite bool
}

func NewBuildCommand() *cobra.Command {
	cmd := &buildCommand{
		engineFactory: func(files fs.FS, data templating.CLIProject) renderer {
			return templating.NewEngine(files, data)
		},
		loader:     templating.Load,
		dataSource: defaultDataSource,
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

			if !fileutils.IsDir(c.outputPath) {
				return &errorhandling.CommandError{
					Err:      fmt.Errorf("output path %q is not a directory", c.outputPath),
					ExitCode: errorhandling.ExitInputError,
					HelpText: "The specified output path is not a directory, please provide a valid directory path.",
				}
			}

			empty, err := fileutils.IsDirEmpty(c.outputPath)
			if err != nil {
				return &errorhandling.CommandError{
					Err:      fmt.Errorf("failed to check output directory %q: %w", c.outputPath, err),
					ExitCode: errorhandling.ExitInputError,
					HelpText: "The output directory must be accessible before building into it.",
				}
			}
			if !empty {
				if c.forceOverwrite {
					logger.Warn("Proceeding with overwrite due to force flag")
				} else {
					return &errorhandling.CommandError{
						Err:      fmt.Errorf("output path %q is not empty", c.outputPath),
						ExitCode: errorhandling.ExitInputError,
						HelpText: "The output directory must be empty before building into it.",
					}
				}
			}

			files, err := c.loader(ctx, category)
			if err != nil {
				return &errorhandling.CommandError{
					Err:      fmt.Errorf("failed to load template resources for category %q: %w", category, err),
					ExitCode: errorhandling.ExitGenericError,
					HelpText: "Failed to load template resources. Please check the category name and try again.",
				}
			}

			data, err := c.dataSource(c.sourceFile)
			if err != nil {
				return &errorhandling.CommandError{
					Err:      fmt.Errorf("failed to resolve template data: %w", err),
					ExitCode: errorhandling.ExitInputError,
					HelpText: "Provide a valid --source YAML file or complete the interactive prompt.",
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
			}).Info("Created full project from template")

			greenFmt := color.New(color.FgGreen).FprintlnFunc()
			greenFmt(cmd.OutOrStdout(), "Built new templated project, happy hacking!")

			return nil
		},
	}

	cmd.Flags().StringVarP(&c.outputPath, "output", "o", ".", "Output directory for the generated project")
	cmd.Flags().StringVarP(&c.sourceFile, "source", "s", "", "Path to a YAML file supplying template data (name, username, author, go-version)")
	cmd.Flags().BoolVarP(&c.forceOverwrite, "force", "f", false, "Force overwrite of existing files in the output directory")
	return cmd
}

// renderer is the minimal interface required by the build command to execute a
// template render. *templating.Engine satisfies this interface.
type renderer interface {
	Render(ctx context.Context, outputPath string) error
}

// loaderFunc loads the embedded FS for a given category.
// Matches templating.Load so it can be swapped in tests.
type loaderFunc func(ctx context.Context, category string) (fs.FS, error)

// templaterFactory constructs a renderer from a loaded FS and template data.
type templaterFactory func(files fs.FS, data templating.CLIProject) renderer
