package tubes

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func initTube(data string) (*Tube, *bytes.Buffer) {
	buf := bytes.NewBufferString(data)
	return NewTube(buf), buf
}

func TestRecvn(t *testing.T) {
	assert := assert.New(t)
	var tube *Tube
	var data []byte
	var err error
	testString := "hello world"

	tube, _ = initTube(testString)
	data, err = tube.RecvN(uint(len(testString)), 0)
	assert.Nil(err)
	assert.EqualValues(testString, data)
}

func BenchmarkRecvn(b *testing.B) {
	testString := strings.Repeat("A", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tube, _ := initTube(testString)
		tube.RecvN(uint(len(testString)), 0)
	}
}

func TestRecvuntil(t *testing.T) {
	assert := assert.New(t)
	var tube *Tube
	var data []byte
	var err error

	tube, _ = initTube("")
	data, err = tube.RecvUntil([]byte{}, false)
	assert.Nil(data)
	assert.EqualError(err, "Empty delimiter")
	data, err = tube.RecvUntil(nil, false)
	assert.Nil(data)
	assert.EqualError(err, "Empty delimiter")
	data, err = tube.RecvUntil(' ', false)
	assert.Nil(data)
	assert.NotNil(err)

	tube, _ = initTube("Hello World!")
	data, err = tube.RecvUntil(' ', false)
	assert.Nil(err)
	assert.EqualValues("Hello ", data)
	data, _ = tube.RecvN(6, 0)
	assert.EqualValues("World!", data)

	tube, _ = initTube("Hello World!")
	data, err = tube.RecvUntil(" Wor", false)
	assert.Nil(err)
	assert.EqualValues("Hello Wor", data)
	data, _ = tube.RecvN(3, 0)
	assert.EqualValues("ld!", data)

	tube, _ = initTube("Hello World!")
	data, err = tube.RecvUntil(" Wor", true)
	assert.Nil(err)
	assert.EqualValues("Hello", data)
	data, _ = tube.RecvN(3, 0)
	assert.EqualValues("ld!", data)

	tube, _ = initTube("Hello|World")
	data, err = tube.RecvUntil('|', true)
	assert.Nil(err)
	assert.EqualValues("Hello", data)
	data, _ = tube.RecvN(5, 0)
	assert.EqualValues("World", data)
}

func BenchmarkRecvuntil(b *testing.B) {
	testString := strings.Repeat("A", 99) + "\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tube, _ := initTube(testString)
		tube.RecvUntil('\n', true)
	}
}

func TestRecvlines(t *testing.T) {
	assert := assert.New(t)
	var tube *Tube
	var lines [][]byte
	var err error

	const delim = "\n"
	const buffer = "Foo\nBar\r\nBaz\n"
	expected := strings.Split(buffer, delim)
	tube, _ = initTube(buffer)
	lines, err = tube.RecvLines(3, false)
	assert.Nil(err)
	for i, line := range lines {
		assert.EqualValues(expected[i], line)
	}

	tube, _ = initTube(buffer)
	lines, err = tube.RecvLines(3, true)
	assert.Nil(err)
	for i, line := range lines {
		assert.EqualValues(expected[i]+"\n", line)
	}

	lines, err = tube.RecvLines(1, false)
	assert.Nil(lines)
	assert.NotNil(err)
}

func TestRecvline(t *testing.T) {
	assert := assert.New(t)
	var tube *Tube
	var data []byte
	var err error

	tube, _ = initTube("Foo\nBar\r\nBaz\n")
	data, err = tube.RecvLine(true)
	assert.Nil(err)
	assert.EqualValues("Foo\n", data)
	data, err = tube.RecvLine(true)
	assert.Nil(err)
	assert.EqualValues("Bar\r\n", data)
	data, err = tube.RecvLine(false)
	assert.Nil(err)
	assert.EqualValues("Baz", data)
}

func BenchmarkRecvline(b *testing.B) {
	testString := strings.Repeat("A", 99) + "\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tube, _ := initTube(testString)
		tube.RecvLine(false)
	}
}

func TestSend(t *testing.T) {
	assert := assert.New(t)
	tube, _ := initTube("")
	var data []byte
	var err error

	data, err = tube.Send("hello")
	assert.Nil(err)
	assert.EqualValues("hello", data)
}

func BenchmarkSend(b *testing.B) {
	testString := strings.Repeat("A", 100)
	tube, buf := initTube(testString)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		tube.Send(testString)
	}
}

func TestSendline(t *testing.T) {
	assert := assert.New(t)
	tube, _ := initTube("")
	var data []byte
	var err error

	assert.Equal(tube.Newline, []byte{'\n'})
	data, err = tube.SendLine("hello")
	assert.Nil(err)
	assert.EqualValues("hello\n", data)
}

func TestSendafter(t *testing.T) {
	assert := assert.New(t)
	tube, _ := initTube("hello world")
	var data []byte
	var err error

	data, err = tube.SendAfter("hello ", "hello")
	assert.Nil(err)
	assert.EqualValues("hello", data)

	data, err = tube.SendAfter('!', []byte{})
	assert.Nil(data)
	assert.NotNil(err)
}

func TestSendlineafter(t *testing.T) {
	assert := assert.New(t)
	tube, _ := initTube("hello world")
	var data []byte
	var err error

	assert.Equal(tube.Newline, []byte{'\n'})
	data, err = tube.SendLineAfter("hello ", "hello")
	assert.Nil(err)
	assert.EqualValues("hello\n", data)
}

func TestSendthen(t *testing.T) {
	assert := assert.New(t)
	tube, _ := initTube("hello world")
	var data []byte
	var err error

	data, err = tube.SendThen("hello ", "hello")
	assert.Nil(err)
	assert.EqualValues("hello ", data)

	data, err = tube.SendThen('!', []byte{})
	assert.Nil(data)
	assert.NotNil(err)
}

func TestSendlinethen(t *testing.T) {
	assert := assert.New(t)
	tube, _ := initTube("hello world")
	var data []byte
	var err error

	assert.Equal(tube.Newline, []byte{'\n'})
	data, err = tube.SendLineThen("hello ", "hello")
	assert.Nil(err)
	assert.EqualValues("hello ", data)
}

func TestParseBytes(t *testing.T) {
	assert := assert.New(t)
	const expected = "Hello World"

	assert.EqualValues([]byte{'!'}, parseBytes(rune('!')))
	assert.EqualValues([]byte{'!'}, parseBytes(byte('!')))
	assert.EqualValues(expected, parseBytes([]byte(expected)))
	assert.EqualValues(expected, parseBytes(expected))
}
