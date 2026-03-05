package templating

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
)

//go:embed resources
var embeddedResources embed.FS

func Load(category string) (fs.FS, error) {
	subDirPath := filepath.Join("resources", category)
	sub, err := fs.Sub(embeddedResources, subDirPath)
	if err != nil {
		return nil, fmt.Errorf("templating failed to load category %q: %w", category, err)
	}
	return sub, nil
}
