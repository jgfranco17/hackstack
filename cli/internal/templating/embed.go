package templating

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/jgfranco17/hackstack/cli/internal/logging"
)

//go:embed resources
var embeddedResources embed.FS

// CLIProject holds the variables required to render the CLI project templates.
// Fields map 1:1 to {{ .Field }} placeholders in `.j2` files.
// YAML tags use kebab-case to match the expected source file format.
type CLIProject struct {
	Name      string `yaml:"name"`
	Username  string `yaml:"username"`
	GoVersion string `yaml:"go-version"`
	Author    string `yaml:"author"`
}

func (d *CLIProject) Validate() error {
	var missing []string
	if d.Name == "" {
		missing = append(missing, "name")
	}
	if d.Username == "" {
		missing = append(missing, "username")
	}
	if d.Author == "" {
		missing = append(missing, "author")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields in source data: %s", strings.Join(missing, ", "))
	}
	return nil
}

func Load(ctx context.Context, category string) (fs.FS, error) {
	logger := logging.FromContext(ctx).WithField("module", "templating")

	category = strings.ToLower(category)
	subDirPath := filepath.Join("resources", category)
	sub, err := fs.Sub(embeddedResources, subDirPath)
	if err != nil {
		return nil, fmt.Errorf("templating failed to load category %q: %w", category, err)
	}

	logger.WithField("category", category).Trace("Loaded template resources")
	return sub, nil
}
