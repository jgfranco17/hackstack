package cmds

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jgfranco17/hackstack/cli/internal/templating"
)

// dataSourceFunc resolves TemplateData from a source file path. When sourceFile
// is empty the implementation should fall back to interactive input.
type dataSourceFunc func(sourceFile string) (templating.CLIProject, error)

// defaultDataSource is the production dataSourceFunc. It reads from a YAML file
// when sourceFile is non-empty, otherwise launches the interactive TUI.
func defaultDataSource(sourceFile string) (templating.CLIProject, error) {
	if sourceFile != "" {
		return loadFromYAMLFile(sourceFile)
	}
	return promptForData()
}

// loadFromYAMLFile reads a YAML file at path and decodes it into CLIProject.
// GoVersion is always overridden with the host runtime version after decoding.
func loadFromYAMLFile(path string) (templating.CLIProject, error) {
	f, err := os.Open(path)
	if err != nil {
		return templating.CLIProject{}, fmt.Errorf("open source file %q: %w", path, err)
	}
	defer f.Close()

	data, err := templating.DataFromSource[templating.CLIProject](f)
	if err != nil {
		return templating.CLIProject{}, fmt.Errorf("parse source file %q: %w", path, err)
	}

	data.GoVersion = runtimeGoVersion()

	if err := data.Validate(); err != nil {
		return templating.CLIProject{}, err
	}
	return *data, nil
}

// runtimeGoVersion returns the host Go version string without the leading "go"
// prefix (e.g. "1.24.0").
func runtimeGoVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}

// promptForData launches an interactive TUI to collect Name, Username, and
// Author. GoVersion is populated automatically from the host runtime.
func promptForData() (templating.CLIProject, error) {
	m, err := tea.NewProgram(newPromptModel(), tea.WithAltScreen()).Run()
	if err != nil {
		return templating.CLIProject{}, fmt.Errorf("TUI error: %w", err)
	}

	result, ok := m.(promptModel)
	if !ok {
		return templating.CLIProject{}, errors.New("unexpected TUI model type")
	}
	if result.cancelled {
		return templating.CLIProject{}, errors.New("cancelled by user")
	}

	return templating.CLIProject{
		Name:      result.inputs[0].Value(),
		Username:  result.inputs[1].Value(),
		Author:    result.inputs[2].Value(),
		GoVersion: runtimeGoVersion(),
	}, nil
}

// promptModel is the bubbletea model for the interactive data-entry TUI.
type promptModel struct {
	inputs    []textinput.Model
	focused   int
	cancelled bool
}

var promptLabels = []string{"Project name", "GitHub username", "Author name"}

func newPromptModel() promptModel {
	inputs := make([]textinput.Model, len(promptLabels))
	for i, label := range promptLabels {
		ti := textinput.New()
		ti.Placeholder = label
		ti.CharLimit = 100
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return promptModel{inputs: inputs}
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter:
			if msg.Type == tea.KeyShiftTab {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused >= len(m.inputs) {
				// All fields filled — submit.
				return m, tea.Quit
			}
			if m.focused < 0 {
				m.focused = 0
			}

			for i := range m.inputs {
				if i == m.focused {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, nil
		}
	}

	// Forward keystrokes to the focused input.
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m promptModel) View() string {
	var b strings.Builder
	b.WriteString("New hackstack project\n\n")
	for i, label := range promptLabels {
		b.WriteString(fmt.Sprintf("%s\n%s\n\n", label, m.inputs[i].View()))
	}
	b.WriteString("Tab / Enter to advance • Shift+Tab to go back • Esc to cancel\n")
	return b.String()
}
