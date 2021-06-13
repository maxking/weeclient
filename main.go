package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"github.com/maxking/weeclient/src/client"

	"github.com/maxking/weeclient/src/weechat"
)

// Set of initial commands sent to weechat.
const (
	authCommand    = `init password=%v`
	initialCommand = `(listbuffers) hdata buffer:gui_buffers(*) number,full_name,short_name,type,nicklist,title,local_variables,
(listlines) hdata buffer:gui_buffers(*)/own_lines/last_line(-%(lines)d)/data date,displayed,prefix,message,buffer
(nicklist) nicklist
sync
`
)

// This requires setting up a relay that is listening at the localhost port 8080.
// If you have a relay running remotely, you can use SSH to essentially replicate
// the same thing by port forwarding.
// sample command::
// ssh -L 8080:localhost:8080 <remote-server>
const relay = "localhost:8080"

func main() {
	conn, err := net.Dial("tcp", relay)
	if err != nil {
		fmt.Printf("Failed to connect to remote relay at %v: %v\n", relay, err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter password for %v\n> ", relay)
	text, _ := reader.ReadString('\n')
	// TODO: handle error.

	_, err = conn.Write([]byte(fmt.Sprintf(authCommand, text)))
	if err != nil {
		fmt.Println("Failed to send auth message")
		os.Exit(1)
	}
	num, err := conn.Write([]byte(initialCommand))
	if err != nil {
		fmt.Println("Failed to send auth message")
		os.Exit(1)
	}
	fmt.Printf("<-- Sending (%v bytes) %v\n", num, string(initialCommand))

	if err != nil {
		fmt.Println("Failed to authenticate with remote weechat relay.")
	}
	fmt.Printf("<-- Sent %v bytes\n", num)

	weeproto := weechat.Protocol{}

	weechan := make(chan *weechat.WeechatMessage)

	// sendchan := make(chan []bytes)

	go func() {
		for {
			// first, read the length of the next message and block on
			msgLen := make([]byte, 4)
			_, err = conn.Read(msgLen)
			if err != nil {
				fmt.Printf("Failed to read message length. %v", err)
			}
			length := int(binary.BigEndian.Uint32(msgLen)) - 4
			// now, read the complete message (msglen - 4 bytes for the length.)
			msg := make([]byte, length)
			_, err = conn.Read(msg)
			if err != nil {
				fmt.Printf("Failed to read message of lenth %v, err: %v", msgLen, err)
			}

			weeMsg, err := weeproto.Decode(append(msgLen, msg...))
			if err != nil {
				fmt.Printf("Failed to decode message from weechat. %v", err)
			} else {
				weechan <- weeMsg
			}
		}
	}()

	// handle sending of message.

	client.TviewStart(weechan)

}
