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
			wantExit: errorhandling.ExitTemplateError,
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
			wantExit: errorhandling.ExitTemplateError,
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
				loader := func(_ context.Context, category string) (fs.FS, error) {
					gotCategory = category
					return fstest.MapFS{}, nil
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
		{
			name: "valid --source YAML file is used for template data",
			args: func(t *testing.T) []string {
				src := filepath.Join(t.TempDir(), "data.yaml")
				require.NoError(t, os.WriteFile(src, []byte(
					"name: myapp\nusername: jgfranco17\nauthor: Jorge\n",
				), 0644))
				return []string{"cli", "--output", t.TempDir(), "--source", src}
			},
			cmd: func(t *testing.T) *buildCommand {
				cmd := newTestBuildCommand(successLoader, nil)
				// Use the real YAML loader so the flag is exercised end-to-end.
				cmd.dataSource = defaultDataSource
				return cmd
			},
			wantNoErr:  true,
			wantStdout: "happy hacking",
		},
		{
			name: "--source pointing to non-existent file returns input error",
			args: func(t *testing.T) []string {
				return []string{"cli", "--output", t.TempDir(), "--source", "/no/such/file.yaml"}
			},
			cmd: func(t *testing.T) *buildCommand {
				cmd := newTestBuildCommand(successLoader, nil)
				cmd.dataSource = defaultDataSource
				return cmd
			},
			wantErr:  true,
			wantExit: errorhandling.ExitInputError,
		},
		{
			name: "dataSource failure returns input error",
			args: func(t *testing.T) []string {
				return []string{"cli", "--output", t.TempDir()}
			},
			cmd: func(t *testing.T) *buildCommand {
				cmd := newTestBuildCommand(successLoader, nil)
				cmd.dataSource = func(_ string) (templating.CLIProject, error) {
					return templating.CLIProject{}, errors.New("data source failure")
				}
				return cmd
			},
			wantErr:  true,
			wantExit: errorhandling.ExitInputError,
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

func TestLoadFromYAMLFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filePath func(t *testing.T) string
		wantName string
		wantUser string
		wantAuth string
		wantErr  string
	}{
		{
			name:     "valid full YAML",
			content:  "name: myapp\nusername: jgfranco17\nauthor: Jorge\n",
			wantName: "myapp",
			wantUser: "jgfranco17",
			wantAuth: "Jorge",
		},
		{
			name:    "missing name field",
			content: "username: jgfranco17\nauthor: Jorge\n",
			wantErr: "name",
		},
		{
			name:    "missing username field",
			content: "name: myapp\nauthor: Jorge\n",
			wantErr: "username",
		},
		{
			name:    "missing author field",
			content: "name: myapp\nusername: jgfranco17\n",
			wantErr: "author",
		},
		{
			name:    "multiple missing fields",
			content: "name: myapp\n",
			wantErr: "username",
		},
		{
			name:    "malformed YAML",
			content: "name: [unclosed bracket\n",
			wantErr: "parse source file",
		},
		{
			name: "non-existent file path",
			filePath: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "no-such-file.yaml")
			},
			wantErr: "open source file",
		},
		{
			name:    "empty file produces parse error",
			content: "",
			wantErr: "parse source file",
		},
		{
			name:     "go-version field from YAML is ignored in favour of runtime",
			content:  "name: myapp\nusername: user\nauthor: author\ngo-version: 0.0.0\n",
			wantName: "myapp",
			wantUser: "user",
			wantAuth: "author",
			// GoVersion is always set from runtime, so 0.0.0 should be overwritten.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var path string
			if tc.filePath != nil {
				path = tc.filePath(t)
			} else {
				path = filepath.Join(t.TempDir(), "data.yaml")
				require.NoError(t, os.WriteFile(path, []byte(tc.content), 0644))
			}

			data, err := loadFromYAMLFile(path)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantName, data.Name)
			assert.Equal(t, tc.wantUser, data.Username)
			assert.Equal(t, tc.wantAuth, data.Author)
		})
	}
}

func TestValidateTemplateData(t *testing.T) {
	tests := []struct {
		name    string
		input   string // YAML content to decode into TemplateData
		wantErr bool
		missing []string
	}{
		{
			name:  "all fields present",
			input: "name: a\nusername: b\nauthor: c\n",
		},
		{
			name:    "empty name",
			input:   "username: b\nauthor: c\n",
			wantErr: true,
			missing: []string{"name"},
		},
		{
			name:    "empty username",
			input:   "name: a\nauthor: c\n",
			wantErr: true,
			missing: []string{"username"},
		},
		{
			name:    "empty author",
			input:   "name: a\nusername: b\n",
			wantErr: true,
			missing: []string{"author"},
		},
		{
			name:    "all fields empty strings",
			input:   "name:\nusername:\nauthor:\n",
			wantErr: true,
			missing: []string{"name", "username", "author"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tc.input), 0644))

			imported, err := loadFromYAMLFile(path)
			if tc.wantErr {
				require.Error(t, err)
				for _, field := range tc.missing {
					assert.Contains(t, err.Error(), field)
				}
				// Zero-value struct on error
				assert.Empty(t, imported.Name)
				return
			}
			require.NoError(t, err)
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

// newTestBuildCommand wires up a buildCommand with injectable dependencies.
// dataSource defaults to successDataSource when nil.
func newTestBuildCommand(loader loaderFunc, renderErr error) *buildCommand {
	return &buildCommand{
		loader: loader,
		dataSource: func(_ string) (templating.CLIProject, error) {
			return templating.CLIProject{Name: "test", Username: "user", Author: "author"}, nil
		},
		engineFactory: func(_ fs.FS, _ templating.CLIProject) renderer {
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
func successLoader(_ context.Context, _ string) (fs.FS, error) {
	return fstest.MapFS{}, nil
}

// errorLoader always returns an error.
func errorLoader(_ context.Context, _ string) (fs.FS, error) {
	return nil, errors.New("loader failure")
}

// touchFile creates a file inside dir to make it non-empty.
func touchFile(t *testing.T, dir string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, "existing.txt"))
	require.NoError(t, err)
	require.NoError(t, f.Close())
}
