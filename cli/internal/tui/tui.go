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

// PromptForCLI launches an interactive TUI to collect project details and returns the
// resulting CLIProject. GoVersion is populated from the host runtime.
func PromptForCLI() (templating.CLIProject, error) {
	m, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run()
	if err != nil {
		return templating.CLIProject{}, fmt.Errorf("TUI error: %w", err)
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

// model is the bubbletea model for the interactive data-entry TUI.
type model struct {
	inputs    []textinput.Model
	focused   int
	cancelled bool
}

var labels = []string{"Project name", "GitHub username", "Author name"}

func newModel() model {
	inputs := make([]textinput.Model, len(labels))
	for i, label := range labels {
		ti := textinput.New()
		ti.Placeholder = label
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

func (m model) View() string {
	var b strings.Builder
	b.WriteString("New hackstack project\n\n")
	for i, label := range labels {
		b.WriteString(fmt.Sprintf("%s\n%s\n\n", label, m.inputs[i].View()))
	}
	b.WriteString("Tab / Enter to advance • Shift+Tab to go back • Esc to cancel\n")
	return b.String()
}
