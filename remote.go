package tubes

import (
	"bufio"
	"errors"
	"net"
	"sort"
	"strconv"
)

var protocols = []string{"tcp", "udp"}
var nProtocols = len(protocols)

// Remote represents a remote connection tube
type Remote struct {
	Tube
	conn net.Conn
}

// NewRemote returns a new process tube
func NewRemote(host string, port int, proto string) (*Remote, error) {
	if sort.SearchStrings(protocols, proto) == nProtocols {
		return nil, errors.New("Invalid protocol")
	}

	conn, err := net.Dial(proto, host+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	remote := &Remote{
		Tube: Tube{
			fd:      bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
			Newline: '\n',
		},
		conn: conn,
	}
	remote.CloseFunc = remote.Close
	return remote, nil
}

// Close terminates the underlying network connection
func (r *Remote) Close() error {
	return r.conn.Close()
}
