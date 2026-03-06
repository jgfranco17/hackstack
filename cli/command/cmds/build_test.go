package cmds

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jgfranco17/hackstack/cli/internal/errorhandling"
	"github.com/jgfranco17/hackstack/cli/internal/logging"
	"github.com/jgfranco17/hackstack/cli/internal/templating"
	"github.com/jgfranco17/hackstack/cli/internal/testutils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       func(t *testing.T) []string
		cmd        func(t *testing.T) *buildCommand
		wantErr    bool
		wantExit   errorhandling.ExitCode
		wantStdout string
		wantNoErr  bool
	}{
		{
			name: "output path does not exist",
			args: func(t *testing.T) []string {
				return []string{"cli", "--output", filepath.Join(t.TempDir(), "no-such-dir")}
			},
			cmd: func(t *testing.T) *buildCommand {
				return newTestBuildCommand(successLoader, nil)
			},
			wantErr:  true,
			wantExit: errorhandling.ExitInputError,
		},
		{
			name: "output path is a file not a directory",
			args: func(t *testing.T) []string {
				path := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(path, []byte("x"), 0644))
				return []string{"cli", "--output", path}
			},
			cmd: func(t *testing.T) *buildCommand {
				return newTestBuildCommand(successLoader, nil)
			},
			wantErr:  true,
			wantExit: errorhandling.ExitInputError,
		},
		{
			name: "non-empty output dir without force flag returns error",
			args: func(t *testing.T) []string {
				dir := t.TempDir()
				touchFile(t, dir)
				return []string{"cli", "--output", dir}
			},
			cmd: func(t *testing.T) *buildCommand {
				return newTestBuildCommand(successLoader, nil)
			},
			wantErr:  true,
			wantExit: errorhandling.ExitInputError,
		},
		{
			name: "non-empty output dir with force flag proceeds",
			args: func(t *testing.T) []string {
				dir := t.TempDir()
				touchFile(t, dir)
				return []string{"cli", "--output", dir, "--force"}
			},
			cmd: func(t *testing.T) *buildCommand {
				return newTestBuildCommand(successLoader, nil)
			},
			wantNoErr:  true,
			wantStdout: "happy hacking",
		},
		{
			name: "loader failure returns generic error",
			args: func(t *testing.T) []string {
				return []string{"cli", "--output", t.TempDir()}
			},
			cmd: func(t *testing.T) *buildCommand {
				return newTestBuildCommand(errorLoader, nil)
			},
			wantErr:  true,
			wantExit: errorhandling.ExitGenericError,
		},
		{
			name: "renderer failure returns generic error",
			args: func(t *testing.T) []string {
				return []string{"cli", "--output", t.TempDir()}
			},
			cmd: func(t *testing.T) *buildCommand {
				return newTestBuildCommand(successLoader, errors.New("render failure"))
			},
			wantErr:  true,
			wantExit: errorhandling.ExitGenericError,
		},
		{
			name: "successful render on empty dir prints success message",
			args: func(t *testing.T) []string {
				return []string{"cli", "--output", t.TempDir()}
			},
			cmd: func(t *testing.T) *buildCommand {
				return newTestBuildCommand(successLoader, nil)
			},
			wantNoErr:  true,
			wantStdout: "happy hacking",
		},
		{
			name: "category argument is forwarded to loader",
			args: func(t *testing.T) []string {
				return []string{"mycat", "--output", t.TempDir()}
			},
			cmd: func(t *testing.T) *buildCommand {
				var gotCategory string
				loader := func(_ context.Context, category string) (fs.FS, templating.DynamicData, error) {
					gotCategory = category
					return fstest.MapFS{}, templating.DynamicData{}, nil
				}
				cmd := newTestBuildCommand(loader, nil)
				t.Cleanup(func() { assert.Equal(t, "mycat", gotCategory) })
				return cmd
			},
			wantNoErr: true,
		},
		{
			name: "missing category argument returns error",
			args: func(t *testing.T) []string { return []string{} },
			cmd: func(t *testing.T) *buildCommand {
				return newTestBuildCommand(successLoader, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := runBuild(t, tc.cmd(t), tc.args(t)...)

			if tc.wantNoErr {
				require.NoError(t, result.RunErr)
			}

			if tc.wantErr {
				require.Error(t, result.RunErr)
			}

			if tc.wantExit != 0 {
				var cmdErr *errorhandling.CommandError
				require.True(t, errors.As(result.RunErr, &cmdErr), "expected a CommandError, got: %v", result.RunErr)
				assert.Equal(t, tc.wantExit, cmdErr.ExitCode)
			}

			if tc.wantStdout != "" {
				assert.True(t, strings.Contains(result.Stdout, tc.wantStdout),
					"expected stdout to contain %q, got: %q", tc.wantStdout, result.Stdout)
			}
		})
	}
}

// mockRenderer implements renderer, returning a fixed error (nil = success).
type mockRenderer struct {
	err error
}

func (m *mockRenderer) Render(_ context.Context, _ string) error {
	return m.err
}

// newTestBuildCommand wires up a buildCommand with injectable loader and renderer.
func newTestBuildCommand(loader loaderFunc, renderErr error) *buildCommand {
	return &buildCommand{
		loader: loader,
		engineFactory: func(_ fs.FS, _ templating.DynamicData) renderer {
			return &mockRenderer{err: renderErr}
		},
	}
}

// buildContext returns a context carrying a silent logger, satisfying
// logging.FromContext without cluttering test output.
func buildContext(t *testing.T) context.Context {
	t.Helper()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logging.AddToContext(context.Background(), logger)
}

// runBuild executes the build cobra command with the supplied args,
// after injecting a logger-carrying context.
func runBuild(t *testing.T, cmd *buildCommand, args ...string) testutils.ExecResult {
	t.Helper()
	cobraCmd := cmd.invoke()
	cobraCmd.SetContext(buildContext(t))
	return testutils.RunCommand(t, cobraCmd, args...)
}

// successLoader is a no-op loader that always returns an empty FS and no error.
func successLoader(_ context.Context, _ string) (fs.FS, templating.DynamicData, error) {
	return fstest.MapFS{}, templating.DynamicData{}, nil
}

// errorLoader always returns an error.
func errorLoader(_ context.Context, _ string) (fs.FS, templating.DynamicData, error) {
	return nil, nil, errors.New("loader failure")
}

// touchFile creates a file inside dir to make it non-empty.
func touchFile(t *testing.T, dir string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, "existing.txt"))
	require.NoError(t, err)
	require.NoError(t, f.Close())
}
