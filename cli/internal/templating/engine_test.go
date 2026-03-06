package templating

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/jgfranco17/hackstack/cli/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContext(t *testing.T) context.Context {
	t.Helper()
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.TraceLevel)
	return logging.AddToContext(t.Context(), logger)
}

func TestEngine_Render(t *testing.T) {
	tests := []struct {
		name        string
		files       fstest.MapFS
		data        CLIProject
		wantFiles   map[string]string // relative output path -> expected content
		unwantFiles []string          // relative output paths that must NOT exist
		wantErr     bool
	}{
		{
			name: "copy file is copied verbatim and extension stripped",
			files: fstest.MapFS{
				"README.md.copy": {Data: []byte("# Hello world")},
			},
			data: CLIProject{},
			wantFiles: map[string]string{
				"README.md": "# Hello world",
			},
			unwantFiles: []string{"README.md.copy"},
		},
		{
			name: "j2 file is rendered and extension stripped",
			files: fstest.MapFS{
				"main.go.j2": {Data: []byte("package {{ .Name }}")},
			},
			data: CLIProject{Name: "mypkg"},
			wantFiles: map[string]string{
				"main.go": "package mypkg",
			},
			unwantFiles: []string{"main.go.j2"},
		},
		{
			name: "nested directory structure is preserved",
			files: fstest.MapFS{
				"cli/cmd/root.go.j2": {Data: []byte("// {{ .Name }}")},
				"cli/main.go.copy":   {Data: []byte("package main")},
			},
			data: CLIProject{Name: "myapp"},
			wantFiles: map[string]string{
				"cli/cmd/root.go": "// myapp",
				"cli/main.go":     "package main",
			},
			unwantFiles: []string{"cli/cmd/root.go.j2", "cli/main.go.copy"},
		},
		{
			name: "multiple template variables are substituted",
			files: fstest.MapFS{
				"go.mod.j2": {Data: []byte("module github.com/{{ .Username }}/{{ .Name }}\n\ngo {{ .GoVersion }}")},
			},
			data: CLIProject{
				Username:  "jgfranco17",
				Name:      "myapp",
				GoVersion: "1.24.0",
			},
			wantFiles: map[string]string{
				"go.mod": "module github.com/jgfranco17/myapp\n\ngo 1.24.0",
			},
		},
		{
			name: "invalid template syntax returns error",
			files: fstest.MapFS{
				"broken.go.j2": {Data: []byte("package {{ .Name")},
			},
			data:    CLIProject{},
			wantErr: true,
		},
		{
			name: "unrecognized extension returns error",
			files: fstest.MapFS{
				"README.md": {Data: []byte("# Hello")},
			},
			data:    CLIProject{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outDir := t.TempDir()
			ctx := testContext(t)

			engine := NewEngine(tc.files, tc.data)
			err := engine.Render(ctx, outDir)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			for relPath, wantContent := range tc.wantFiles {
				fullPath := filepath.Join(outDir, filepath.FromSlash(relPath))
				got, readErr := os.ReadFile(fullPath)
				require.NoError(t, readErr, "expected output file %q to exist", relPath)
				assert.Equal(t, wantContent, string(got), "content mismatch for %q", relPath)
			}

			for _, relPath := range tc.unwantFiles {
				fullPath := filepath.Join(outDir, filepath.FromSlash(relPath))
				assert.NoFileExists(t, fullPath)
			}
		})
	}
}
