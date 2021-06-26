package main

import (
	"bufio"
	"fmt"
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

	conn := weechat.WeechatConnFactory(weechat.WebsocketConnection, relay, "/weechat", true)
	if err := conn.Connect(); err != nil {
		fmt.Printf("Failed to connect to remote relay at %v: %v\n", relay, err)
		os.Exit(1)
	}

	fmt.Printf("Enter password for %v\n> ", relay)
	text, _ := reader.ReadString('\n')
	// TODO: handle error.

	err := conn.Write([]byte(fmt.Sprintf(authCommand, text)))
	if err != nil {
		fmt.Println("Failed to send auth message")
		os.Exit(1)
	}

	err = conn.Write([]byte(initialCommand))
	if err != nil {
		fmt.Println("Failed to send auth message")
		os.Exit(1)
	}
	weeproto := weechat.Protocol{}

	// Channel to process incoming message and passing it on to terminal ui.
	// Goroutine runs in a loop and listens to messages from weechat relay.
	weechan := make(chan *weechat.WeechatMessage)
	// handle receiving of message.
	go func() {
		for {
			// first, read the length of the next message and block on
			// msgLen := make([]byte, 4)
			// _, err = conn.Read(msgLen)
			// if err != nil {
			// 	fmt.Printf("Failed to read message length. %v", err)
			// }
			// length := int(binary.BigEndian.Uint32(msgLen)) - 4
			// // now, read the complete message (msglen - 4 bytes for the length.)
			// msg := make([]byte, length)
			// _, err = conn.Read(msg)
			// if err != nil {
			// 	fmt.Printf("Failed to read message of lenth %v, err: %v", msgLen, err)
			// }
			msg, err := conn.Read()
			if err != nil {
				fmt.Printf("Failed to read message over websocket:%v\n", err)
			}
			weeMsg, err := weeproto.Decode(msg)
			if err != nil {
				weechan <- &weechat.WeechatMessage{
					Msgid:  "error",
					Object: weechat.WeechatObject{ObjType: "error", Value: err},
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
			if err = conn.Write([]byte(sendmsg)); err != nil {
				// do something if failed to send message.
			}
		}
	}()

	// Start the terminal app.
	client.TviewStart(weechan, sendchan)
}
