package fileutils

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	file, err := os.Create(path.Join(tmpDir, "file"))
	require.NoError(t, err)
	defer file.Close()

	testCases := []struct {
		description string
		testPath    string
		isDir       bool
	}{
		{
			description: "Valid Directory",
			testPath:    tmpDir,
			isDir:       true,
		},
		{
			description: "Unknown Directory",
			testPath:    path.Join(tmpDir, "unknown"),
			isDir:       false,
		},
		{
			description: "Valid File",
			testPath:    path.Join(tmpDir, "file"),
			isDir:       false,
		},
		{
			description: "Unknown File",
			testPath:    path.Join(tmpDir, "unknownFile"),
			isDir:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.isDir, IsDir(tc.testPath))
		})
	}
}

func TestIsDirEmpty(t *testing.T) {
	testCases := []struct {
		description string
		setup       func(t *testing.T) string
		wantEmpty   bool
		wantErr     bool
	}{
		{
			description: "empty directory",
			setup:       func(t *testing.T) string { return t.TempDir() },
			wantEmpty:   true,
		},
		{
			description: "directory with a file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				f, err := os.Create(path.Join(dir, "file.txt"))
				require.NoError(t, err)
				f.Close()
				return dir
			},
			wantEmpty: false,
		},
		{
			description: "directory with a subdirectory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				require.NoError(t, os.Mkdir(path.Join(dir, "sub"), 0755))
				return dir
			},
			wantEmpty: false,
		},
		{
			description: "non-existent path returns error",
			setup:       func(t *testing.T) string { return path.Join(t.TempDir(), "does-not-exist") },
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dir := tc.setup(t)
			got, err := IsDirEmpty(dir)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantEmpty, got)
		})
	}
}
