package tui

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jgfranco17/hackstack/cli/internal/templating"
)

// PromptForCLI launches an interactive TUI that walks the user through each
// project field one at a time and returns the resulting CLIProject.
// GoVersion is populated from the host runtime.
func PromptForCLI() (templating.CLIProject, error) {
	m, err := tea.NewProgram(newModel()).Run()
	if err != nil {
		return templating.CLIProject{}, fmt.Errorf("failed to load TUI: %w", err)
	}

	result, ok := m.(model)
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
		GoVersion: strings.TrimPrefix(runtime.Version(), "go"),
	}, nil
}

// fieldDef holds the display label for a single prompt field.
type fieldDef struct {
	label string
}

var fields = []fieldDef{
	{"Project name"},
	{"GitHub username"},
	{"Author name"},
}

// model is the bubbletea model for the step-by-step data-entry TUI.
type model struct {
	inputs    []textinput.Model
	step      int
	cancelled bool
}

func newModel() model {
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.label
		ti.CharLimit = 100
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return model{inputs: inputs}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyEnter:
			m.step++
			if m.step >= len(m.inputs) {
				// All fields confirmed — submit.
				return m, tea.Quit
			}
			m.inputs[m.step-1].Blur()
			m.inputs[m.step].Focus()
			return m, nil
		}
	}

	// Forward keystrokes to the current input.
	var cmd tea.Cmd
	m.inputs[m.step], cmd = m.inputs[m.step].Update(msg)
	return m, cmd
}

func (m model) View() string {
	// bubbletea calls View once more after tea.Quit is returned; guard against
	// the step being out of range at that point.
	if m.step >= len(fields) {
		return ""
	}
	var b strings.Builder
	for i := 0; i < m.step; i++ {
		fmt.Fprintf(&b, "%s: %s\n", fields[i].label, m.inputs[i].Value())
	}
	if m.step > 0 {
		b.WriteString("\n")
	}
	f := fields[m.step]
	fmt.Fprintf(&b, "%s (%d/%d)\n", f.label, m.step+1, len(fields))
	b.WriteString(m.inputs[m.step].View())
	b.WriteString("\n\nEnter to confirm • Ctrl+C to cancel\n")
	return b.String()
}
