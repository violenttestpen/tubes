package tubes

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/sync/errgroup"
)

var errEmptyDelim = errors.New("Empty delimiter")

const bufSize = 64

// Reader represents a Tube reader interface
type Reader interface {
	Recvn(uint, int) ([]byte, error)
	Recvuntil(interface{}, bool) ([]byte, error)
	Recvlines(uint, bool) ([][]byte, error)
	Recvline(bool) ([]byte, error)
	Recvall() ([]byte, error)
}

// Writer represents a Tube writer interface
type Writer interface {
	Send(interface{}) ([]byte, error)
	Sendline(interface{}) ([]byte, error)
	Sendlineafter(interface{}, interface{}) ([]byte, error)
	Sendthen(interface{}, interface{}) ([]byte, error)
	Sendlinethen(interface{}, interface{}) ([]byte, error)
}

// ReadWriter represents a generic Tube interface
type ReadWriter interface {
	Reader
	Writer

	// Clean() []byte
	Interactive() error
}

// Tube represents a wrapped I/O pipeline
type Tube struct {
	fd        *bufio.ReadWriter
	Newline   byte
	CloseFunc func() error
}

// NewTube creates a new tube struct
func NewTube(fd io.ReadWriter) *Tube {
	return &Tube{
		fd:      bufio.NewReadWriter(bufio.NewReader(fd), bufio.NewWriter(fd)),
		Newline: '\n',
	}
}

// RecvN reads n bytes from the tube
func (t *Tube) RecvN(numb uint, timeout int) ([]byte, error) {
	buf := make([]byte, numb, numb)
	_, err := io.ReadFull(t.fd, buf)
	return buf, err
}

// RecvUntil reads until it receives the delimiter
func (t *Tube) RecvUntil(delims interface{}, drop bool) ([]byte, error) {
	delim := parseBytes(delims)
	if len(delim) == 0 {
		return nil, errEmptyDelim
	}

	received := make([]byte, 0)
	for {
		data, err := t.fd.ReadBytes(delim[0])
		if err != nil {
			return nil, err
		}
		received = append(received, data...)
		if data, err = t.fd.Peek(len(delim) - 1); bytes.Equal(data, delim[1:]) {
			t.RecvN(uint(len(data)), 0)
			received = append(received, delim[1:]...)
			break
		}
	}

	index := bytes.Index(received, delim)
	if !drop {
		index += len(delim)
	}
	return received[:index], nil
}

// RecvLines reads up till `n` lines
func (t *Tube) RecvLines(numLines uint, keepends bool) ([][]byte, error) {
	var err error
	lines := make([][]byte, numLines, numLines)
	for i := uint(0); i < numLines; i++ {
		lines[i], err = t.RecvLine(keepends)
		if err != nil {
			return nil, err
		}
	}
	return lines, nil
}

// RecvLine reads until a newline is received
func (t *Tube) RecvLine(keepends bool) ([]byte, error) {
	data, err := t.fd.ReadBytes(t.Newline)
	if keepends {
		return data, err
	}
	return data[:len(data)-1], err
}

// RecvAll reads until the tube is closed
func (t *Tube) RecvAll() ([]byte, error) {
	return io.ReadAll(t.fd)
}

// Send writes data to tube
func (t *Tube) Send(buf interface{}) ([]byte, error) {
	data := parseBytes(buf)
	reader := bytes.NewReader(data)
	nn, err := io.CopyN(t.fd, reader, int64(reader.Len()))
	if err != nil {
		return nil, err
	}
	return data[:nn], err
}

// SendLine writes data ending with a newline to tube
func (t *Tube) SendLine(buf interface{}) ([]byte, error) {
	data := parseBytes(buf)
	return t.Send(append(data, t.Newline))
}

// SendAfter reads until delimiter before writing data to tube
func (t *Tube) SendAfter(delim, data interface{}) ([]byte, error) {
	if _, err := t.RecvUntil(delim, false); err != nil {
		return nil, err
	}
	return t.Send(data)
}

// SendLineAfter reads until delimiter before writing data ending with a newline to tube
func (t *Tube) SendLineAfter(delim, buf interface{}) ([]byte, error) {
	data := parseBytes(buf)
	return t.SendAfter(delim, append(data, t.Newline))
}

// SendThen writing data to tube then reads until delimiter
func (t *Tube) SendThen(delim, data interface{}) ([]byte, error) {
	if _, err := t.Send(data); err != nil {
		return nil, err
	}
	return t.RecvUntil(delim, false)
}

// SendLineThen writing data ending with a newline to tube then reads until delimiter
func (t *Tube) SendLineThen(delim, buf interface{}) ([]byte, error) {
	data := parseBytes(buf)
	return t.SendThen(delim, append(data, t.Newline))
}

// Clean removes all the buffered data from a tube
func (t *Tube) Clean() {
	t.fd.Discard(t.fd.Reader.Buffered())
}

// Interactive forwards all inputs between the program and the service
func (t *Tube) Interactive() error {
	fmt.Println("Switching to interactive mode...")
	grp, ctx := errgroup.WithContext(context.Background())

	grp.Go(func() error {
		_, err := io.Copy(os.Stdout, t.fd)
		fmt.Println("Got EOF while reading in interactive")
		return err
	})

	grp.Go(func() error {
		stdin := bufio.NewScanner(os.Stdin)
		for stdin.Scan() {
			select {
			case <-ctx.Done():
				return t.CloseFunc()
			default:
				input := stdin.Text()
				if _, err := t.SendLine(input); err != nil {
					return err
				}
			}
		}
		return t.CloseFunc()
	})

	return grp.Wait()
}

func parseBytes(buf interface{}) (data []byte) {
	switch d := buf.(type) {
	case rune:
		data = append([]byte(nil), byte(d))
	case byte:
		data = append([]byte(nil), d)
	case []byte:
		data = d
	case string:
		data = []byte(d)
	}
	return
}
