package templating

import (
	"context"
	"io/fs"

	"github.com/jgfranco17/hackstack/cli/internal/logging"
)

type DynamicData map[string]any

type Engine struct {
	Files fs.FS
	Data  DynamicData
}

func NewEngine(files fs.FS, data DynamicData) *Engine {
	return &Engine{
		Files: files,
		Data:  data,
	}
}

func (e *Engine) Render(ctx context.Context, outputPath string) error {
	logger := logging.FromContext(ctx).WithField("module", "templating")
	logger.Trace("Rendering template files")
	// TODO: Implement the actual rendering logic here
	return nil
}
