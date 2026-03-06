package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModel_InitialState(t *testing.T) {
	m := newModel()

	require.Len(t, m.inputs, len(labels))
	assert.Equal(t, 0, m.focused)
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

func TestModel_View(t *testing.T) {
	m := newModel()
	view := m.View()

	assert.Contains(t, view, "New hackstack project")
	for _, label := range labels {
		assert.Contains(t, view, label)
	}
	assert.Contains(t, view, "Tab")
	assert.Contains(t, view, "Esc")
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

func TestModel_Update_Tab_AdvancesFocus(t *testing.T) {
	m := newModel()
	assert.Equal(t, 0, m.focused)

	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})

	result := out.(model)
	assert.Equal(t, 1, result.focused)
	assert.False(t, result.cancelled)
	assert.Nil(t, cmd)
}

func TestModel_Update_ShiftTab_RetractsFocus(t *testing.T) {
	m := newModel()
	m.focused = 1

	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

	result := out.(model)
	assert.Equal(t, 0, result.focused)
	assert.Nil(t, cmd)
}

func TestModel_Update_ShiftTab_ClampsToZero(t *testing.T) {
	m := newModel()
	// focused is already 0; shift-tab must not go below 0.
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})

	result := out.(model)
	assert.Equal(t, 0, result.focused)
	assert.Nil(t, cmd)
}

func TestModel_Update_Enter_AdvancesFocus(t *testing.T) {
	m := newModel()

	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	result := out.(model)
	assert.Equal(t, 1, result.focused)
	assert.Nil(t, cmd)
}

func TestModel_Update_Enter_OnLastField_Quits(t *testing.T) {
	m := newModel()
	m.focused = len(m.inputs) - 1

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.True(t, isQuitCmd(t, cmd), "entering on the last field should quit")
}

func TestModel_Update_Tab_OnLastField_Quits(t *testing.T) {
	m := newModel()
	m.focused = len(m.inputs) - 1

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})

	assert.True(t, isQuitCmd(t, cmd), "tab on the last field should quit")
}

func TestModel_Update_FocusShift_UpdatesInputFocusState(t *testing.T) {
	m := newModel()
	require.True(t, m.inputs[0].Focused())

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	result := out.(model)

	assert.False(t, result.inputs[0].Focused(), "previous input should lose focus")
	assert.True(t, result.inputs[1].Focused(), "next input should gain focus")
}

func TestModel_Update_Rune_ForwardedToFocusedInput(t *testing.T) {
	m := newModel()
	// Type a character into the focused (0th) input.
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	result := out.(model)
	assert.Contains(t, result.inputs[0].Value(), "a")
}

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
