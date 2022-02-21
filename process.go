package tubes

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"time"
)

// Process represents a process tube
type Process struct {
	Tube
	process *os.Process
}

// NewProcess returns a new process tube
func NewProcess(argv ...string) (*Process, error) {
	cmd := exec.Command(argv[0], argv[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go io.Copy(os.Stderr, stderr)

	process := &Process{
		Tube: Tube{
			fd: bufio.NewReadWriter(
				bufio.NewReader(io.MultiReader(stdout, stderr)),
				bufio.NewWriter(stdin),
			),
			Newline: '\n',
		},
		process: cmd.Process,
	}
	process.CloseFunc = process.Close
	return process, nil
}

// Close terminates the underlying network connection
func (p *Process) Close() error {
	done := make(chan error, 1)
	go func() {
		_, err := p.process.Wait()
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-time.Tick(1 * time.Second):
		return p.process.Kill()
	}
}
