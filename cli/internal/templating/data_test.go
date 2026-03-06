package templating

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataFromSource(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CLIProject
		wantErr bool
	}{
		{
			name:  "all fields decoded correctly",
			input: "name: myapp\nusername: jgfranco17\nauthor: John Doe\ngo-version: 1.24.0\n",
			want: CLIProject{
				Name:      "myapp",
				Username:  "jgfranco17",
				Author:    "John Doe",
				GoVersion: "1.24.0",
			},
		},
		{
			name:  "partial fields decoded, missing fields are zero values",
			input: "name: myapp\n",
			want:  CLIProject{Name: "myapp"},
		},
		{
			name:  "unknown fields are silently ignored",
			input: "name: myapp\nunknown-key: ignored\n",
			want:  CLIProject{Name: "myapp"},
		},
		{
			name:    "empty reader returns error",
			input:   "",
			wantErr: true,
		},
		{
			name:    "malformed YAML returns error",
			input:   "name: [unclosed\n",
			wantErr: true,
		},
		{
			name:    "non-YAML content returns error",
			input:   "{ this is not > yaml : at all",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DataFromSource[CLIProject](strings.NewReader(tc.input))

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.want, *got)
		})
	}
}
