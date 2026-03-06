package cmds

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			// GoVersion is always the host runtime version, never empty.
			assert.NotEmpty(t, data.GoVersion)
			if tc.content != "" && tc.wantErr == "" {
				assert.NotEqual(t, "0.0.0", data.GoVersion, "runtime version should override YAML value")
			}
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
