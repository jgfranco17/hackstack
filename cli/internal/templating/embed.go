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

type DynamicData map[string]any

func Load(ctx context.Context, category string) (fs.FS, DynamicData, error) {
	logger := logging.FromContext(ctx).WithField("module", "templating")

	data := make(DynamicData)
	category = strings.ToLower(category)
	switch category {
	case "cli":
		data = DynamicData{
			"Name": "myapp",
		}
	}

	subDirPath := filepath.Join("resources", category)
	sub, err := fs.Sub(embeddedResources, subDirPath)
	if err != nil {
		return nil, nil, fmt.Errorf("templating failed to load category %q: %w", category, err)
	}

	logger.WithField("category", category).Trace("Loaded template resources")
	return sub, data, nil
}
