// Package exe contains helpers to help running commands
package exe

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Run runs a command and return stderr as error
func Run(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	stdout, stderr, err := execCmd(cmd)

	if err != nil && stderr != "" {
		return stdout, errors.New(stderr) //nolint:goerr113 // the error is dynamically generated at runtime
	}

	return stdout, err
}

func execCmd(cmd *exec.Cmd) (stdout, stderr string, err error) {
	// we pipe stderr to get the error message if something goes wrong
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("could not pipe stderr: %w", err)
	}
	// we pipe stdout to get the output of the script
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("could not pipe stddout: %w", err)
	}

	// We start the command
	if err = cmd.Start(); err != nil {
		return "", "", err
	}

	// we read all stderr to get the error message (if any)
	stderrByte, err := io.ReadAll(stderrReader)
	if err != nil {
		return "", "", fmt.Errorf("could not read stderr: %w", err)
	}
	// we read all stdout to get the output of the script (if any)
	stdoutByte, err := io.ReadAll(stdoutReader)
	if err != nil {
		return "", "", fmt.Errorf("could not read stdout: %w", err)
	}

	stdout = strings.TrimSuffix(string(stdoutByte), "\n")
	stderr = strings.TrimSuffix(string(stderrByte), "\n")

	return stdout, stderr, cmd.Wait()
}
