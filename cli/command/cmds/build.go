package cmds

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jgfranco17/hackstack/cli/internal/errorhandling"
	"github.com/jgfranco17/hackstack/cli/internal/fileutils"
	"github.com/jgfranco17/hackstack/cli/internal/logging"
	"github.com/jgfranco17/hackstack/cli/internal/templating"
	"github.com/jgfranco17/hackstack/cli/internal/tui"
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

// dataSourceFunc resolves TemplateData from a source file path. When sourceFile
// is empty the implementation should fall back to interactive input.
type dataSourceFunc func(sourceFile string) (templating.CLIProject, error)

// defaultDataSource is the production dataSourceFunc. It reads from a YAML file
// when sourceFile is non-empty, otherwise launches the interactive TUI.
func defaultDataSource(sourceFile string) (templating.CLIProject, error) {
	if sourceFile != "" {
		return loadFromYAMLFile(sourceFile)
	}
	return promptForData()
}

// loadFromYAMLFile reads a YAML file at path and decodes it into CLIProject.
// GoVersion is always overridden with the host runtime version after decoding.
func loadFromYAMLFile(path string) (templating.CLIProject, error) {
	f, err := os.Open(path)
	if err != nil {
		return templating.CLIProject{}, fmt.Errorf("open source file %q: %w", path, err)
	}
	defer f.Close()

	data, err := templating.DataFromSource[templating.CLIProject](f)
	if err != nil {
		return templating.CLIProject{}, fmt.Errorf("parse source file %q: %w", path, err)
	}

	if data.GoVersion == "" {
		data.GoVersion = runtimeGoVersion()
	}

	if err := data.Validate(); err != nil {
		return templating.CLIProject{}, err
	}
	return *data, nil
}

// runtimeGoVersion returns the host Go version string without the leading "go"
// prefix (e.g. "1.24.0").
func runtimeGoVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}

// promptForData launches an interactive TUI to collect Name, Username, and
// Author. GoVersion is populated automatically from the host runtime.
func promptForData() (templating.CLIProject, error) {
	return tui.PromptForCLI()
}
