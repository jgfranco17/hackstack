package entrypoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"

	"github.com/jgfranco17/hackstack/cli/command"
	"github.com/jgfranco17/hackstack/cli/internal/errorhandling"
)

const (
	commandName = "hackstack"
)

type ProjectMetadata struct {
	Author      string `json:"author"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

func readMetadata(rawData []byte) (ProjectMetadata, error) {
	var projectMetadata ProjectMetadata
	if err := json.Unmarshal(rawData, &projectMetadata); err != nil {
		return ProjectMetadata{}, fmt.Errorf("Error unmarshaling metadata: %v\n", err)
	}
	return projectMetadata, nil
}

func Run(metadata []byte) {
	defer func() {
		if r := recover(); r != nil {
			printError("Application %s crashed: %v", commandName, r)
			os.Exit(errorhandling.ExitPanicError.Int())
		}
	}()

	projectMetadata, err := readMetadata(metadata)
	if err != nil {
		printError("Error reading metadata: %v", err)
		os.Exit(errorhandling.ExitInputError.Int())
	}

	command, err := command.New(command.RootCommandOptions{
		Name:        commandName,
		Description: projectMetadata.Description,
		Version:     projectMetadata.Version,
	})
	if err != nil {
		printError("Error creating command: %v", err)
		os.Exit(errorhandling.ExitRuntimeError.Int())
	}

	var exitCode int
	if err := command.Execute(); err != nil {
		exitCode = handlerExecError(err)
	} else {
		exitCode = 0
	}

	os.Exit(exitCode)
}

func handlerExecError(err error) int {
	var cmdErr *errorhandling.CommandError
	if ok := errors.As(err, &cmdErr); ok {
		fmt.Println(cmdErr.String())
		return cmdErr.ExitCode.Int()
	}
	fmt.Printf("An unexpected error occurred: %v\n", err)
	return errorhandling.ExitRuntimeError.Int()
}

func printError(format string, args ...any) {
	redBoldFmt := color.New(color.FgRed).Add(color.Bold).FprintfFunc()
	message := fmt.Sprintf(format, args...)
	redBoldFmt(os.Stderr, "[FATAL] %s\n", message)
}
