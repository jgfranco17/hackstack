package errorhandling

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandError(t *testing.T) {
	mockErr := CommandError{
		Err:      errors.New("mock error"),
		ExitCode: ExitRuntimeError,
		HelpText: "This is some help text.",
	}

	require.Error(t, &mockErr, "CommandError should be accepted as error")
	assert.Error(t, mockErr.Unwrap())
	assert.Equal(t, "mock error", mockErr.Error())
	assert.Equal(t, "This is some help text.", mockErr.HelpText)
	assert.Contains(t, mockErr.String(), "[ERROR]")
	assert.Contains(t, mockErr.String(), "[HELP]")
}
