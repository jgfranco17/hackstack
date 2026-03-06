package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isQuitCmd returns true when cmd produces a tea.QuitMsg, which is how
// bubbletea signals a program exit.
func isQuitCmd(t *testing.T, cmd tea.Cmd) bool {
	t.Helper()
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestNewModel_InitialState(t *testing.T) {
	m := newModel()

	require.Len(t, m.inputs, len(fields))
	assert.Equal(t, 0, m.step)
	assert.False(t, m.cancelled)
	assert.True(t, m.inputs[0].Focused(), "first input should be focused by default")
	for i := 1; i < len(m.inputs); i++ {
		assert.False(t, m.inputs[i].Focused(), "input %d should not be focused by default", i)
	}
}

func TestModel_Init(t *testing.T) {
	m := newModel()
	cmd := m.Init()
	assert.NotNil(t, cmd, "Init should return a non-nil Cmd (textinput.Blink)")
}

func TestModel_View_ShowsCurrentFieldAndProgress(t *testing.T) {
	m := newModel()
	view := m.View()

	assert.Contains(t, view, fields[0].label)
	assert.Contains(t, view, fmt.Sprintf("1/%d", len(fields)))
	assert.Contains(t, view, "Enter")
	assert.NotContains(t, view, fields[1].label, "only the current field should be visible")
}

func TestModel_View_UpdatesAfterStepAdvance(t *testing.T) {
	m := newModel()
	m.inputs[0].SetValue("myapp")

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := out.(model)
	view := result.View()

	assert.Contains(t, view, fields[1].label, "current field label should be shown")
	assert.Contains(t, view, fmt.Sprintf("2/%d", len(fields)))
	assert.Contains(t, view, fields[0].label, "completed field label should appear in summary")
	assert.Contains(t, view, "myapp", "completed field value should appear in summary")
}

func TestModel_View_ShowsAllCompletedFields(t *testing.T) {
	m := newModel()
	m.inputs[0].SetValue("myapp")
	m.inputs[1].SetValue("jgfranco17")
	m.step = 2
	view := m.View()

	assert.Contains(t, view, "myapp")
	assert.Contains(t, view, "jgfranco17")
	assert.Contains(t, view, fields[2].label, "current field label should be shown")
}

func TestModel_Update_CtrlC_SetsCancel(t *testing.T) {
	m := newModel()

	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	result := out.(model)
	assert.True(t, result.cancelled)
	assert.True(t, isQuitCmd(t, cmd))
}

func TestModel_Update_Esc_SetsCancel(t *testing.T) {
	m := newModel()

	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	result := out.(model)
	assert.True(t, result.cancelled)
	assert.True(t, isQuitCmd(t, cmd))
}

func TestModel_Update_Enter_AdvancesStep(t *testing.T) {
	m := newModel()

	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	result := out.(model)
	assert.Equal(t, 1, result.step)
	assert.False(t, result.cancelled)
	assert.Nil(t, cmd)
}

func TestModel_Update_Enter_UpdatesFocusState(t *testing.T) {
	m := newModel()
	require.True(t, m.inputs[0].Focused())

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := out.(model)

	assert.False(t, result.inputs[0].Focused(), "completed input should lose focus")
	assert.True(t, result.inputs[1].Focused(), "next input should gain focus")
}

func TestModel_Update_Enter_OnLastField_Quits(t *testing.T) {
	m := newModel()
	m.step = len(m.inputs) - 1

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.True(t, isQuitCmd(t, cmd), "confirming the last field should quit")
}

func TestModel_View_AfterSubmit_DoesNotPanic(t *testing.T) {
	// bubbletea calls View once more after tea.Quit is returned; step will be
	// out of range at that point.
	m := newModel()
	m.step = len(fields)

	assert.NotPanics(t, func() { m.View() })
	assert.Empty(t, m.View())
}

func TestModel_Update_Rune_ForwardedToCurrentInput(t *testing.T) {
	m := newModel()
	// Clear the pre-filled default then type a new value.
	m.inputs[0].SetValue("")
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	result := out.(model)
	assert.Contains(t, result.inputs[0].Value(), "x")
}

func TestModel_Update_Rune_NotForwardedToOtherInputs(t *testing.T) {
	m := newModel()
	m.inputs[1].SetValue("")
	m.inputs[2].SetValue("")

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	result := out.(model)

	assert.Empty(t, result.inputs[1].Value(), "non-active input should not receive keystrokes")
	assert.Empty(t, result.inputs[2].Value(), "non-active input should not receive keystrokes")
}
