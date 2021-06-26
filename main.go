package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/maxking/weeclient/src/client"

	"github.com/maxking/weeclient/src/weechat"
)

// Set of initial commands sent to weechat.
const (
	authCommand    = `init password=%v`
	initialCommand = `(listbuffers) hdata buffer:gui_buffers(*) number,full_name,short_name,type,nicklist,title,local_variables,
(listlines) hdata buffer:gui_buffers(*)/own_lines/last_line(-15)/data date,displayed,prefix,message,buffer
sync
`
)

// This requires setting up a relay that is listening at the localhost port 8080.
// If you have a relay running remotely, you can use SSH to essentially replicate
// the same thing by port forwarding.
// sample command::
// ssh -L 8080:localhost:8080 <remote-server>

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Relay \n> ")
	relay, _ := reader.ReadString('\n')
	relay = strings.TrimSuffix(relay, "\n")

	conn, err := net.Dial("tcp", relay)
	if err != nil {
		fmt.Printf("Failed to connect to remote relay at %v: %v\n", relay, err)
		os.Exit(1)
	}
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

	// Channel to process incoming message and passing it on to terminal ui.
	// Goroutine runs in a loop and listens to messages from weechat relay.
	weechan := make(chan *weechat.WeechatMessage)
	// handle receiving of message.
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
				// fmt.Printf("Failed to decode message from weechat. %v", err)
				weechan <- &weechat.WeechatMessage{
					Msgid:  "error",
					Object: weechat.WeechatObject{"error", err},
				}
			} else {
				weechan <- weeMsg
			}
		}
	}()

	// channel to send message. message is received from terminal ui and sent to remote
	// server in the goroutine.
	sendchan := make(chan string)
	// handle sending of message.
	go func() {
		for sendmsg := range sendchan {
			conn.Write([]byte(sendmsg))
		}
	}()

	// Start the terminal app.
	client.TviewStart(weechan, sendchan)
}
