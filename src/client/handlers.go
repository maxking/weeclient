package client

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/maxking/weeclient/src/weechat"
	"github.com/rivo/tview"
)

// *******************************************
// Methods for HandleWeechatMessage interface.
// *******************************************

// Handler (listbufffers) msg that we send at the first boot
// to receive all the currently opened buffers. Pass down each
// buffer to be handled individually.
func (tv *TerminalView) HandleListBuffers(buflist map[string]*weechat.WeechatBuffer) {
	for ptr, buf := range buflist {
		tv.HandleBufferOpened(ptr, buf)
	}
}

// Handles a new buffer opened. This is called several times during the
// startup when the application boots up.
func (tv *TerminalView) HandleBufferOpened(ptr string, buf *weechat.WeechatBuffer) {
	// Add a new item to the List widget.
	tv.bufferList.List.AddItem(fmt.Sprintf("%v", buf.FullName), buf.FullName, 0, nil)

	// Create views for the main chat buffer.
	bufferView := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true).
		SetDynamicColors(true).
		SetText(buf.TitleStr(true) + buf.GetLines(true))

	input := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorGray).
		SetFieldTextColor(tcell.ColorWhite).
		SetPlaceholderTextColor(tcell.ColorWhiteSmoke).
		SetPlaceholder("Type here...")

	// Set handlers for key events in the input box.
	// 1. Enter -> Send message and clear box
	// 2. Esc -> clear box
	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			sendobj := weechat.WeechatSendMessage{
				Message: input.GetText(),
				Buffer:  buf.FullName}
			tv.sendchan <- &sendobj
			input.SetText("")
		case tcell.KeyEscape:
			input.SetText("")
		}

	})

	// Grid for the buffer view with a top row with all the chat and then
	// a bottom input box, of length 1.
	layout := tview.NewGrid().
		SetRows(-1, 1).
		SetBorders(false).
		AddItem(bufferView, 0, 0, 1, 1, 0, 0, false).
		AddItem(input, 1, 0, 1, 1, 0, 0, true)

	// Buffer is a weechat buffer object, which includes the WeechatBuffer object
	// and all the widgets associated with a single buffer window. In future, this
	// will grow to add more widgets like nicklist for example.
	buffer := &Buffer{WeechatBuffer: buf, Input: input, Chat: bufferView}
	tv.bufferList.Buffers[ptr] = buffer

	// The main biffer view page.
	bufferView.SetTitle(buf.FullName)

	// TextView() doesn't auto-refresh when text is added to it. Make sure to
	// refresh it manually.
	bufferView.SetChangedFunc(func() { tv.app.Draw() })

	// Add keybindings to switch between input and chat for focus.
	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			tv.app.SetFocus(bufferView)
		case tcell.KeyCtrlI:
			tv.app.SetFocus(input)
		}
		return event
	})

	// Finally, add a new page for the current buffer. This makes it
	// easy to switch between buffers but switching between pages. It
	// uses a Grid as the primary
	tv.pages.AddPage(fmt.Sprintf("page-%v", buf.FullName),
		layout,
		true,
		// Only the core.weechat buffer is visible at first.
		buffer.FullName == "core.weechat",
	)

	tv.buffers[buf.FullName] = bufferView
}

func (tv *TerminalView) HandleNickList(msg *weechat.WeechatMessage) {

}

// Handle a _buffer_line_added event from Weechat server.
func (tv *TerminalView) HandleLineAdded(line *weechat.WeechatLine) {
	buf := tv.bufferList.Buffers[line.Buffer]
	buf.Lines = append(buf.Lines, line)
	// Also, add the message to the current view.
	if bufView, ok := tv.buffers[buf.FullName]; ok {
		bufView.Write([]byte(line.ToString(true)))
	}
}

// Default handler which handles all the unhandled messages.
func (tv *TerminalView) Default(msg *weechat.WeechatMessage) {

}
