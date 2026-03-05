package command

import (
	"bytes"
	"testing"

	"github.com/jgfranco17/lazyfile/cli/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommandOptions_validate(t *testing.T) {
	tests := []struct {
		name    string
		options RootCommandOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			options: RootCommandOptions{
				Name:        "",
				Description: "A test CLI application",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty version",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "",
			},
			wantErr: true,
		},
		{
			name: "empty description allowed",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		options RootCommandOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "invalid options - no name",
			options: RootCommandOptions{
				Name:        "",
				Description: "A test CLI application",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "invalid options - no version",
			options: RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := New(tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cli)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cli)
				assert.NotNil(t, cli.root)
				assert.Equal(t, tt.options.Name, cli.root.Use)
				assert.Equal(t, tt.options.Version, cli.root.Version)
				assert.Equal(t, tt.options.Description, cli.root.Short)
			}
		})
	}
}

func TestVerbosityLevels(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedLevel logrus.Level
	}{
		{
			name:          "no verbosity flag",
			args:          []string{"test"},
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "single verbose flag",
			args:          []string{"-v", "test"},
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "double verbose flag",
			args:          []string{"-vv", "test"},
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "triple verbose flag",
			args:          []string{"-vvv", "test"},
			expectedLevel: logrus.TraceLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanupCalled bool

			options := RootCommandOptions{
				Name:        "testcli",
				Description: "A test CLI application",
				Version:     "1.0.0",
				CleanupFuncs: []func(){
					func() { cleanupCalled = true },
				},
			}

			cli, err := New(options)
			require.NoError(t, err)

			var actualLevel logrus.Level
			testCmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {
					logger := logging.FromContext(cmd.Context())
					actualLevel = logger.GetLevel()
				},
			}
			cli.RegisterCommands([]*cobra.Command{testCmd})

			var buf bytes.Buffer
			cli.root.SetOut(&buf)
			cli.root.SetErr(&buf)
			cli.root.SetArgs(tt.args)

			exitCode := cli.Execute()
			assert.Equal(t, 0, exitCode)
			assert.Equal(t, tt.expectedLevel, actualLevel)

			cli.Cleanup()
			assert.True(t, cleanupCalled)
		})
	}
}

func TestExecute_HelpFlag(t *testing.T) {
	options := RootCommandOptions{
		Name:        "testcli",
		Description: "A test CLI application",
		Version:     "1.0.0",
	}

	cli, err := New(options)
	require.NoError(t, err)

	var buf bytes.Buffer
	cli.root.SetOut(&buf)
	cli.root.SetErr(&buf)
	cli.root.SetArgs([]string{"--help"})

	exitCode := cli.Execute()
	assert.Equal(t, 0, exitCode)

	output := buf.String()
	assert.Contains(t, output, "A test CLI application")
}

func TestExecute_VersionFlag(t *testing.T) {
	options := RootCommandOptions{
		Name:        "testcli",
		Description: "A test CLI application",
		Version:     "1.0.0",
	}

	cli, err := New(options)
	require.NoError(t, err)

	var buf bytes.Buffer
	cli.root.SetOut(&buf)
	cli.root.SetErr(&buf)
	cli.root.SetArgs([]string{"--version"})

	exitCode := cli.Execute()
	assert.Equal(t, 0, exitCode)

	output := buf.String()
	assert.Contains(t, output, "1.0.0")
}

func TestCleanup_WithoutExecute(t *testing.T) {
	var cleanupCalled bool
	options := RootCommandOptions{
		Name:    "testcli",
		Version: "1.0.0",
		CleanupFuncs: []func(){
			func() { cleanupCalled = true },
		},
	}

	cli, err := New(options)
	require.NoError(t, err)

	assert.NotPanics(t, func() { cli.Cleanup() })
	assert.True(t, cleanupCalled)
}

func TestCleanup_Order(t *testing.T) {
	var order []string

	options := RootCommandOptions{
		Name:    "testcli",
		Version: "1.0.0",
		CleanupFuncs: []func(){
			func() { order = append(order, "user1") },
			func() { order = append(order, "user2") },
		},
	}

	cli, err := New(options)
	require.NoError(t, err)

	original := cli.cleanups[0]
	cli.cleanups[0] = func() {
		order = append(order, "internal")
		original()
	}

	cli.Cleanup()
	assert.Equal(t, []string{"internal", "user1", "user2"}, order)
}
