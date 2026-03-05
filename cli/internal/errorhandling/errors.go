package errorhandling

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type ExitCode int

func (e ExitCode) Int() int {
	return int(e)
}

const (
	ExitSuccess       ExitCode = 0
	ExitGenericError  ExitCode = 1
	ExitTemplateError ExitCode = 2
)

type CommandError struct {
	Err      error
	ExitCode ExitCode
	HelpText string
}

func (e *CommandError) Error() string {
	return e.Err.Error()
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

func (e *CommandError) String() string {
	redBoldFmt := color.New(color.FgRed).Add(color.Bold).SprintfFunc()

	var renderedMessage strings.Builder
	renderedMessage.WriteString("\n")
	renderedMessage.WriteString(redBoldFmt("[ERROR] Hackstack internal error: %s\n", e.Err.Error()))
	if e.HelpText != "" {
		renderedMessage.WriteString(fmt.Sprintf("[HELP] %s", e.HelpText))
	}
	return renderedMessage.String()
}
