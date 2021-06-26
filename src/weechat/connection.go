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
	Read() ([]byte, error)
	Write([]byte) error
	Connect() error
}

type ConnectionType int

const (
	WebsocketConnection ConnectionType = iota
	RelayConnection
)

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

type WeechatWebsocketConn struct {
	URL  *url.URL
	conn websocket.Conn
}

func NewWebsocketConn(host string, path string, ssl bool) *WeechatWebsocketConn {
	url := &url.URL{Host: host, Path: path}
	if ssl {
		url.Scheme = "wss"
	} else {
		url.Scheme = "ws"
	}
	return &WeechatWebsocketConn{URL: url}
}

func (w *WeechatWebsocketConn) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(w.URL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to remote relay at %v: %v",
			w.URL.String(), err)
	}
	w.conn = *conn
	return nil
}

func (w *WeechatWebsocketConn) Write(data []byte) error {
	err := w.conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	return nil
}

func (w *WeechatWebsocketConn) Read() ([]byte, error) {
	_, msg, err := w.conn.ReadMessage()
	return msg, err
}

type WeechatRelayConn struct {
	URL  string
	conn net.Conn
}

func NewRelayConn(url string) *WeechatRelayConn {
	return &WeechatRelayConn{URL: url}
}

func (w *WeechatRelayConn) Connect() error {
	conn, err := net.Dial("tcp", w.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to relay: %v", err)
	}
	w.conn = conn
	return nil
}

func (w *WeechatRelayConn) Write(data []byte) error {
	_, err := w.conn.Write(data)
	return err
}

func (w *WeechatRelayConn) Read() ([]byte, error) {
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
