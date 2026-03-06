package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jgfranco17/hackstack/cli/command/cmds"
	"github.com/jgfranco17/hackstack/cli/internal/errorhandling"
	"github.com/jgfranco17/hackstack/cli/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CLI is a struct that represents the command-line interface of the application.
type CLI struct {
	root     *cobra.Command
	cleanups []func()
}

// ContextModifiers is a function type that takes a context and returns
// a modified context. This can be used to add additional values to the
// context for downstream consumption.
type ContextModifiers func(ctx context.Context) context.Context

// RootCommandOptions defines the options for creating a new CLI instance.
type RootCommandOptions struct {
	// Name is the name of the root command, i.e. the namespace used to invoke the CLI.
	Name string

	// Description is a brief description of the root command.
	// This will be displayed in the help message.
	Description string

	// Version is the version of the root command.
	// This will be displayed in the --version flag.
	Version string

	// CleanupFuncs are functions that will be called when the CLI is cleaned up.
	// This can be used to clean up resources, such as closing database connections
	// or stopping background goroutines.
	CleanupFuncs []func()
}

// validate checks if the required fields in RootCommandOptions are set.
func (options RootCommandOptions) validate() error {
	if options.Name == "" {
		return errors.New("root command must have name")
	}
	if options.Version == "" {
		return errors.New("root command must have version")
	}
	return nil
}

// New creates a new instance of CLI with the provided options.
func New(options RootCommandOptions) (*CLI, error) {
	if err := options.validate(); err != nil {
		return nil, err
	}

	var verbosity int
	root := &cobra.Command{
		Use:     options.Name,
		Version: options.Version,
		Short:   options.Description,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var level logrus.Level
			switch verbosity {
			case 1:
				level = logrus.InfoLevel
			case 2:
				level = logrus.DebugLevel
			case 3:
				level = logrus.TraceLevel
			default:
				level = logrus.WarnLevel
			}

			logger := logging.New(cmd.ErrOrStderr(), level)
			ctx := logging.AddToContext(cmd.Context(), logger)

			ctx, cancel := context.WithCancel(ctx)
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				select {
				case sig := <-c:
					logger.WithField("signal", sig).Infof("Received signal, exiting")
					cancel()
				case <-ctx.Done():
					// Context was canceled, exit goroutine.
				}
			}()

			cmd.SetContext(ctx)
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity (up to -vvv)")

	// Internal cleanup (stop) is prepended so signal notifications are released
	// before user-provided cleanup functions run.
	allCleanups := make([]func(), 0, 1+len(options.CleanupFuncs))
	allCleanups = append(allCleanups, options.CleanupFuncs...)

	root.AddCommand(cmds.NewBuildCommand())

	return &CLI{
		root:     root,
		cleanups: allCleanups,
	}, nil
}

// RegisterCommands registers new commands with the CLI.
func (cr *CLI) RegisterCommands(commands []*cobra.Command) {
	cr.root.AddCommand(commands...)
}

// Cleanup releases resources held by the CLI.
func (cr *CLI) Cleanup() {
	if len(cr.cleanups) > 0 {
		for _, cleanup := range cr.cleanups {
			cleanup()
		}
	}
}

// Execute executes the root command.
func (cr *CLI) Execute() int {
	if err := cr.root.Execute(); err != nil {
		var cmdErr *errorhandling.CommandError
		if ok := errors.As(err, &cmdErr); ok {
			fmt.Println(cmdErr.String())
			return cmdErr.ExitCode.Int()
		} else {
			fmt.Printf("An unexpected error occurred: %v\n", err)
			return errorhandling.ExitGenericError.Int()
		}
	}
	return errorhandling.ExitSuccess.Int()
}
