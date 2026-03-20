package templating

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"text/template"

	"golang.org/x/sync/errgroup"

	"github.com/jgfranco17/hackstack/cli/internal/fileutils"
	"github.com/jgfranco17/hackstack/cli/internal/logging"
)

type Engine struct {
	Files fs.FS
	Data  CLIProject
}

func NewEngine(files fs.FS, data CLIProject) *Engine {
	return &Engine{
		Files: files,
		Data:  data,
	}
}

func (e *Engine) Render(ctx context.Context, outputPath string) error {
	logger := logging.FromContext(ctx).WithField("module", "templating")

	var count atomic.Int64
	g, ctx := errgroup.WithContext(ctx)

	walker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk error at %q: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}

		destPath := filepath.Join(outputPath, filepath.FromSlash(path))

		var work func() error
		switch {
		case strings.HasSuffix(path, ".j2"):
			destPath = strings.TrimSuffix(destPath, ".j2")
			work = func() error {
				logger.WithField("file", path).Trace("Rendering from template")
				return renderTemplate(e.Files, path, destPath, e.Data)
			}
		case strings.HasSuffix(path, ".copy"):
			destPath = strings.TrimSuffix(destPath, ".copy")
			work = func() error {
				logger.WithField("file", path).Trace("Copying file")
				return fileutils.CopyFile(e.Files, path, destPath)
			}
		default:
			return fmt.Errorf("unrecognized resource extension for %q", path)
		}

		count.Add(1)
		g.Go(work)
		return nil
	}

	if err := fs.WalkDir(e.Files, ".", walker); err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	logger.WithField("fileCount", count.Load()).Debug("Completed render")
	return nil
}

func renderTemplate(fsys fs.FS, srcPath, destPath string, data CLIProject) error {
	content, err := fs.ReadFile(fsys, srcPath)
	if err != nil {
		return fmt.Errorf("read template %q: %w", srcPath, err)
	}

	tmpl, err := template.New(srcPath).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template %q: %w", srcPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("create dirs for %q: %w", destPath, err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file %q: %w", destPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template %q: %w", srcPath, err)
	}
	return nil
}
