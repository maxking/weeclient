// Testing CLI for Weechat relay protocol.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/maxking/weeclient/src/color"
	"github.com/maxking/weeclient/src/weechat"
)

// Set of initial commands sent to weechat.
const (
	authCommand    = `init password=%v`
	initialCommand = `(listbuffers) hdata buffer:gui_buffers(*) number,full_name,short_name,type,nicklist,title,local_variables,
(listlines) hdata buffer:gui_buffers(*)/own_lines/last_line(-%(lines)d)/data date,displayed,prefix,message,buffer
sync
`
)

func init() {
	weechat.DebugPrint = true
}

// var relay = "localhost:8080"

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Relay \n> ")
	relay, _ := reader.ReadString('\n')
	relay = strings.TrimSuffix(relay, "\n")

	// conn, err := net.Dial("tcp", relay)
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
	// fmt.Printf("<-- Sending (%v bytes) %v\n", num, string(initialCommand))

	// if err != nil {
	// 	fmt.Println("Failed to authenticate with remote weechat relay.")
	// }
	// fmt.Printf("<-- Sent %v bytes\n", num)

	weeproto := weechat.Protocol{}

	handler := TerminalPrintHandler{}
	// weechan := make(chan *weechat.WeechatMessage)

	go func() {
		for {
			// first, read the length of the next message and block on
			// msgLen := make([]byte, 4)
			// _, err = conn.ReadMessage(msgLen)
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
				fmt.Printf("Failed to decode message from weechat. %v\n", err)
			} else {
				weechat.HandleMessage(weeMsg, &handler)
			}
		}
	}()

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		// text = strings.Replace(text, "\n", "", -1)
		if strings.Compare("quit\n", text) == 0 {
			break
		}
		if text != "\n" {
			if err = conn.Write([]byte(text)); err != nil {
				fmt.Println("Failed to send message.")
			}
			fmt.Printf(color.Green+"<-- %v\n"+color.Reset, text)
		}

	}
}

type TerminalPrintHandler struct {
	// prints all events to terminal.
}

func (mh *TerminalPrintHandler) HandleListBuffers(buflist map[string]*weechat.WeechatBuffer) {
	fmt.Println(color.Red + "Buffers: " + color.Reset)
	for _, buf := range buflist {
		fmt.Printf(color.Red+"\t%v\n"+color.Reset, buf.FullName)
	}

}

func (mh *TerminalPrintHandler) HandleListLines() {
	// noop.
}

func (mh *TerminalPrintHandler) HandleNickList(buffer string, nicks []*weechat.WeechatNick) {
	fmt.Printf("Nicklist %v: %v\n", buffer, nicks)
}

func (mh *TerminalPrintHandler) HandleLineAdded(line *weechat.WeechatLine) {
	fmt.Printf(color.Cyan+"%: %v \n"+color.Reset, line.Buffer, line.ToString(false))
}

func (mh *TerminalPrintHandler) Default(msg *weechat.WeechatMessage) {
	fmt.Printf(color.Gray+"Msgid: %v size: %v\n"+color.Reset, msg.Msgid, msg.Size)
}

func (mh *TerminalPrintHandler) Debug(msg string) {
	fmt.Println(msg)
}
