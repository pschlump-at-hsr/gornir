package task

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
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

	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		fmt.Println(err.Error())
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

	wr := make(chan []byte, 10)

	go func() {
		for {
			select {
			case d := <-wr:
				_, err := stdin.Write(d)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for {
			if tkn := scanner.Scan(); tkn {
				rcv := scanner.Bytes()

				raw := make([]byte, len(rcv))
				copy(raw, rcv)

				RemoteChainedCommandsResults{Stdin: nil, Stdout: scanner.Bytes(), Stderr: nil}.String()
			} else {
				if scanner.Err() != nil {
					RemoteChainedCommandsResults{Stdin: nil, Stdout: scanner.Bytes(), Stderr: []byte(scanner.Err().Error())}.String()
				} else {
					fmt.Println("io.EOF")
				}
				return
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			RemoteChainedCommandsResults{Stdin: nil, Stdout: nil, Stderr: scanner.Bytes()}.String()
		}
	}()

	for _, command := range t.Commands {
		_, err := stdin.Write([]byte(command))
		if err != nil {
			return RemoteChainedCommandsResults{}, errors.Wrap(err, "failed to write to stdin")
		}
	}
	session.Shell()
	time.Sleep(time.Second * 1)
	return RemoteChainedCommandsResults{}, nil
}
