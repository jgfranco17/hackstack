package templating

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIProject_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       CLIProject
		wantErr     bool
		wantMissing []string
	}{
		{
			name:  "all fields present",
			input: CLIProject{Name: "myapp", Username: "user", Author: "Author"},
		},
		{
			name:        "missing name",
			input:       CLIProject{Username: "user", Author: "Author"},
			wantErr:     true,
			wantMissing: []string{"name"},
		},
		{
			name:        "missing username",
			input:       CLIProject{Name: "myapp", Author: "Author"},
			wantErr:     true,
			wantMissing: []string{"username"},
		},
		{
			name:        "missing author",
			input:       CLIProject{Name: "myapp", Username: "user"},
			wantErr:     true,
			wantMissing: []string{"author"},
		},
		{
			name:        "all required fields missing",
			input:       CLIProject{},
			wantErr:     true,
			wantMissing: []string{"name", "username", "author"},
		},
		{
			name:    "GoVersion missing does not affect validation",
			input:   CLIProject{Name: "myapp", Username: "user", Author: "Author"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()

			if tc.wantErr {
				require.Error(t, err)
				for _, field := range tc.wantMissing {
					assert.Contains(t, err.Error(), field)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestLoad_ValidCategory(t *testing.T) {
	ctx := testContext(t)
	sub, err := Load(ctx, "cli")

	require.NoError(t, err)
	require.NotNil(t, sub)

	// Confirm the returned FS is non-empty.
	entries, err := fs.ReadDir(sub, ".")
	require.NoError(t, err)
	assert.NotEmpty(t, entries, "loaded FS should contain template files")
}

func TestLoad_CategoryIsCaseInsensitive(t *testing.T) {
	ctx := testContext(t)

	lower, err := Load(ctx, "cli")
	require.NoError(t, err)

	upper, err := Load(ctx, "CLI")
	require.NoError(t, err)

	// Both should resolve to the same set of top-level entries.
	lowerEntries, err := fs.ReadDir(lower, ".")
	require.NoError(t, err)

	upperEntries, err := fs.ReadDir(upper, ".")
	require.NoError(t, err)

	require.Len(t, upperEntries, len(lowerEntries))
	for i := range lowerEntries {
		assert.Equal(t, lowerEntries[i].Name(), upperEntries[i].Name())
	}
}
