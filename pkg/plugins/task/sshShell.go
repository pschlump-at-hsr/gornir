package task

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/pschlump-at-hsr/gornir/pkg/gornir"
	"github.com/pschlump-at-hsr/gornir/pkg/plugins/connection"

	"github.com/pkg/errors"
)

// RemoteChainedCommands will open a new Session on an already opened ssh connection and execute the given command
type RemoteChainedCommands struct {
	Commands []string             // Command to execute
	Meta     *gornir.TaskMetadata // Task metadata
}

// Metadata returns the task metadata
func (t *RemoteChainedCommands) Metadata() *gornir.TaskMetadata {
	return t.Meta
}

// RemoteChainedCommandsResults is the result of calling RemoteChainedCommands
type RemoteChainedCommandsResults struct {
	Stdout []byte // Stdout written by the command
	Stderr []byte // Stderr written by the command
	Stdin  []byte // Stdin  input commands into the running session
}

// String implemente Stringer interface
func (r RemoteChainedCommandsResults) String() string {
	return fmt.Sprintf("- stdin: %s\n  - stdout: %s\n  - stderr: %s", r.Stdin, r.Stdout, r.Stderr)
}

// Run runs a command on a remote device via ssh
func (t *RemoteChainedCommands) Run(ctx context.Context, logger gornir.Logger, host *gornir.Host) (gornir.TaskInstanceResult, error) {
	conn, err := host.GetConnection("ssh")
	if err != nil {
		return RemoteChainedCommandsResults{}, errors.Wrap(err, "failed to retrieve connection")
	}
	sshConn := conn.(*connection.SSH)

	session, err := sshConn.Client.NewSession()
	if err != nil {
		return RemoteChainedCommandsResults{}, errors.Wrap(err, "failed to create session")
	}

	var (
		stdin          io.WriteCloser
		stdout, stderr io.Reader
	)

	stdin, err = session.StdinPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	stdout, err = session.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
	}

	stderr, err = session.StderrPipe()
	if err != nil {
		fmt.Println(err.Error())
	}
	scannerOut := bufio.NewScanner(stdout)
	scannerErr := bufio.NewScanner(stderr)

	if err := session.Shell(); err != nil {
		return RemoteChainedCommandsResults{}, errors.Wrap(err, "failed to execute command")
	}

	for _, command := range t.Commands {
		_, err := stdin.Write([]byte(command))
		if err != nil {
			return RemoteChainedCommandsResults{}, errors.Wrap(err, "failed to write to stdin")
		}

		for {
			tknOut := scannerOut.Scan()
			tknErr := scannerErr.Scan()
			if tknOut || tknErr {

				RemoteChainedCommandsResults{Stdin: []byte(command), Stdout: scannerOut.Bytes(), Stderr: scannerErr.Bytes()}.String()

			} else {
				if scannerOut.Err() != nil {
					return RemoteChainedCommandsResults{}, errors.Wrap(scannerOut.Err(), "failed to retrieve connection")
				} else {
					time.Sleep(time.Millisecond * 100)
				}

			}
		}

	}

	return RemoteChainedCommandsResults{}, nil
}
