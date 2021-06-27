package weechat

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/url"

	"github.com/gorilla/websocket"
)

// Weechat connection interface represents a connection to the weechat
// relay. There can be more than one way to connect to the weechat
// relay server and this interface wraps that complexity into a single
// interface.
type WeechatConn interface {
	// Read a single weechat message.
	Read() ([]byte, error)
	// Write the bytes to the established connection.
	Write([]byte) error
	// Connect to the relay and save the connection state internally.
	Connect() error
}

// A connection type represents different ways in which we can connect
// to a remote relay.
type ConnectionType int

const (
	// Connection over a http websocket. This is used when the relay sits
	// behind a reverse proxy server like Nginx. Supports SSL (and
	// defaults without a way to disable that yet)
	WebsocketConnection ConnectionType = iota
	// Connect directly to relay over tcp.
	RelayConnection
)

// WeechatConnFactory return a conn object following WeechatConn interface.
// This wraps various types of connections that we support and abstracts
// the implementation details on how those connection types read a single
// message from weechat relay.
// Some parameters are currently unused for certain types of connections,
// for example, relay conn doesn't use the path and ssl parameters. Since,
// Golang doesn't provide a good way to use optional parameters, we have
// to pass in zero value for the unused parameters.
func WeechatConnFactory(connType ConnectionType, url string, path string, ssl bool) WeechatConn {
	switch connType {
	case WebsocketConnection:
		return NewWebsocketConn(url, path, ssl)
	case RelayConnection:
		// relay connection doesn't take the "path" and doesn't
		// support ssl for now.
		return NewRelayConn(url)
	default:
		panic(fmt.Sprintf("unsupported connType %v", connType))

	}
}

// WeechatWebsocetConn object connects to Weechat over a HTTP Websocket
// so that it can talk to relays behind reverse proxies.
type websocketConn struct {
	URL  *url.URL
	conn websocket.Conn
}

// Create a new WeechatWebsocketConn object.
func NewWebsocketConn(host string, path string, ssl bool) *websocketConn {
	url := &url.URL{Host: host, Path: path}
	if ssl {
		url.Scheme = "wss"
	} else {
		url.Scheme = "ws"
	}
	return &websocketConn{URL: url}
}

func (w *websocketConn) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(w.URL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to remote relay at %v: %v",
			w.URL.String(), err)
	}
	w.conn = *conn
	return nil
}

func (w *websocketConn) Write(data []byte) error {
	err := w.conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	return nil
}

func (w *websocketConn) Read() ([]byte, error) {
	_, msg, err := w.conn.ReadMessage()
	return msg, err
}

// This connects directly to the weechat relay over tcp without any
// http layer in between.
type relayConn struct {
	URL  string
	conn net.Conn
}

// Create a new WeechatRelayConn instance.
func NewRelayConn(url string) *relayConn {
	return &relayConn{URL: url}
}

func (w *relayConn) Connect() error {
	conn, err := net.Dial("tcp", w.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to relay: %v", err)
	}
	w.conn = conn
	return nil
}

func (w *relayConn) Write(data []byte) error {
	_, err := w.conn.Write(data)
	return err
}

// This is an interesting implemtnation than the websocket one.
// net.Conn.Read() needs the exact bytes to read and in order
// to read a single message completely, we first read the 4byte
// length, convert it to int32 and then use it to read the whole
// message. Then we combine the length and message object and
// return the bytes.
func (w *relayConn) Read() ([]byte, error) {
	msgLen := make([]byte, 4)
	_, err := w.conn.Read(msgLen)
	if err != nil {
		return nil, fmt.Errorf("failed to read message length. %v", err)
	}
	length := int(binary.BigEndian.Uint32(msgLen)) - 4
	// now, read the complete message (msglen - 4 bytes for the length.)
	msg := make([]byte, length)
	_, err = w.conn.Read(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to read message of lenth %v, err: %v", msgLen, err)
	}
	msgLen = append(msgLen, msg...)
	return msgLen, nil
}
