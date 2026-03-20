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
	"github.com/sirupsen/logrus"
)

const (
	extensionTemplate = ".j2"
	extensionRawCopy  = ".copy"
)

// Engine is the entity responsible for rendering embedded template
// files with provided source data.
type Engine struct {
	Files fs.FS
	Data  CLIProject
}

// NewEngine creates a new templating engine instance with the provided
// embedded files and source data.
func NewEngine(files fs.FS, data CLIProject) *Engine {
	return &Engine{
		Files: files,
		Data:  data,
	}
}

// Render processes the embedded template files and writes the output to the specified directory.
// It walks through all files in the embedded FS, rendering templates and copying raw files as
// needed. The function returns an error if any step of the rendering process fails.
func (e *Engine) Render(ctx context.Context, outputPath string) error {
	logger := logging.FromContext(ctx).WithField("module", "templating")

	var count atomic.Int64
	g, ctx := errgroup.WithContext(ctx)

	walker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk error at %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}

		destPath := filepath.Join(outputPath, filepath.FromSlash(path))

		var work func() error
		switch {
		case strings.HasSuffix(path, extensionTemplate):
			destPath = strings.TrimSuffix(destPath, extensionTemplate)
			work = func() error {
				logger.WithFields(logrus.Fields{
					"source":      path,
					"destination": destPath,
				}).Trace("Rendering from template")
				return renderTemplate(e.Files, path, destPath, e.Data)
			}
		case strings.HasSuffix(path, extensionRawCopy):
			destPath = strings.TrimSuffix(destPath, extensionRawCopy)
			work = func() error {
				logger.WithFields(logrus.Fields{
					"source":      path,
					"destination": destPath,
				}).Trace("Copying raw file")
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
