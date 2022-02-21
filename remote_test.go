package tubes

import (
	"io"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const host = "127.0.0.1"
const proto = "tcp"

var port string

func RunFakeServer() <-chan int {
	ready := make(chan int, 1)
	go func(ready chan<- int) {
		listener, err := net.Listen(proto, host+":0")
		if err != nil {
			panic(err)
		}
		defer listener.Close()
		port = strings.Split(listener.Addr().String(), ":")[1]

		ready <- 1
		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}
			go func(conn net.Conn) { io.Copy(conn, conn) }(conn)
		}
	}(ready)
	return ready
}

func TestNewRemote(t *testing.T) {
	assert := assert.New(t)
	<-RunFakeServer()

	var r *Remote
	var err error
	portNum, err := strconv.Atoi(port)
	r, err = NewRemote(host, portNum, proto)
	assert.NotNil(r)
	assert.NoError(err)

	const expected = "hello world"
	d, err := r.SendLine([]byte(expected))
	assert.EqualValues(d, expected+"\n")
	assert.NoError(err)

	d, err = r.RecvLine(false)
	assert.EqualValues(d, expected)
	assert.NoError(err)

	err = r.CloseFunc()
	assert.NoError(err)

	r, err = NewRemote("127.0.0.1", portNum, "xxx")
	assert.Nil(r)
	assert.Error(err)

	r, err = NewRemote("127.0.0.1", 0, "tcp")
	assert.Nil(r)
	assert.Error(err)
}
