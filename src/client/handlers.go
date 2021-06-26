package client

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/maxking/weeclient/src/color"
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
	tv.bufferList.List.AddItem(fmt.Sprintf("%v", buf.FullName), "", 0, nil)

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
			sendobj := &weechat.WeechatSendMessage{
				Message: input.GetText(),
				Buffer:  buf.FullName}
			tv.sendchan <- sendobj.String()
			input.SetText("")
		case tcell.KeyEscape:
			input.SetText("")
		}

	})

	// nick list of the buffer.
	nicklist := tview.NewList().ShowSecondaryText(false)

	// Grid for the buffer view with a top row with all the chat and then
	// a bottom input box, of length 1.
	layout := tview.NewGrid().
		SetRows(-1, 1).
		SetColumns(-1, 15).
		SetBorders(false).
		AddItem(bufferView, 0, 0, 1, 1, 0, 0, false).
		AddItem(input, 1, 0, 1, 1, 0, 0, true).
		AddItem(nicklist, 0, 1, 1, 1, 0, 0, false)

	// Buffer is a weechat buffer object, which includes the WeechatBuffer object
	// and all the widgets associated with a single buffer window. In future, this
	// will grow to add more widgets like nicklist for example.
	buffer := &Buffer{WeechatBuffer: buf, Input: input, Chat: bufferView, NickList: nicklist}
	tv.bufferList.Buffers[ptr] = buffer

	// The main biffer view page.
	bufferView.SetTitle(buf.FullName)

	// TextView() doesn't auto-refresh when text is added to it. Make sure to
	// refresh it manually.
	bufferView.SetChangedFunc(func() {
		indices := tv.bufferList.List.FindItems(buf.FullName, "", true, false)
		if len(indices) == 0 {
			tv.Debug(fmt.Sprintf("Failed to find the buffer for the name %v count: %v\n", buf.FullName, len(indices)))
			return
		}
		// If more than one matched, find the exact match, otherwise, foud
		// represent the index of the buffer.
		var found int
		if len(indices) > 1 {
			found = indices[0]
			for _, index := range indices {
				main, _ := tv.bufferList.List.GetItemText(index)
				if color.RemoveColor(main) == buf.FullName {
					found = index
					break
				}
			}
		} else {
			found = indices[0]
		}
		// Don't do anything if this is the current buffer. Usually,
		// ChangedFunc is called when some text is added and also
		// when the current view is selected as the current item. Hence
		// we need to check for ourselves.
		current := tv.bufferList.List.GetCurrentItem()
		if current == found {
			tv.app.Draw()
			return
		}
		tv.app.QueueUpdateDraw(func() {
			// Mark the buffer color.
			tv.bufferList.List.SetItemText(
				found,
				fmt.Sprintf("[%v]%v[%v]",
					color.UnreadColor, buf.FullName, color.DefaultColor), "")
		})

	})

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

func (tv *TerminalView) HandleNickList(buffer string, nicks []*weechat.WeechatNick) {
	// handle nicklist.
	buf := tv.bufferList.Buffers[buffer]
	if buf != nil {
		if buf.NickList.GetItemCount() != 0 {
			buf.NickList.Clear()
		}
		for _, nick := range nicks {
			if !nick.Group && nick.Level == 0 {
				buf.NickList.AddItem(nick.String(), "", 0, nil)
				tv.app.Draw()
			}
		}
	} else {
		tv.Debug(fmt.Sprintf("Failed to add nicks %v to buffer %v as buffer == nil\n", nicks, buffer))
	}
}

// Handle a _buffer_line_added event from Weechat server.
func (tv *TerminalView) HandleLineAdded(line *weechat.WeechatLine) {
	buf := tv.bufferList.Buffers[line.Buffer]
	buf.Lines = append(buf.Lines, line)
	// Also, add the message to the current view.
	if bufView, ok := tv.buffers[buf.FullName]; ok {
		bufView.Write([]byte("\n" + line.ToString(true)))
	}
}

// Default handler which handles all the unhandled messages.
func (tv *TerminalView) Default(msg *weechat.WeechatMessage) {
	tv.Debug(
		fmt.Sprintf("Uhandled message MsgId: %v ObjType: %v Size: %v Value: %v\n",
			msg.Msgid, msg.Type, msg.Size, msg.Object.Value))
	if msg.Type == weechat.OBJ_HDA {
		tv.Debug(msg.Object.Value.(weechat.WeechatHdaValue).DebugPrint())
	}
}

// DebugPrint will create a new "debug" buffer and write messages to it.
func (tv *TerminalView) Debug(message string) {
	var debug *tview.TextView
	var ok bool
	debug, ok = tv.buffers["debug"]
	if !ok {
		debug = tv.creatDebugBuffer()
	}
	debug.Write([]byte(message))
}

// Helper function to create a new debug buffer to store messages.
func (tv *TerminalView) creatDebugBuffer() *tview.TextView {
	// create a new debugging buffer that is local only and only prints.
	tv.bufferList.List.AddItem("[red]debug[white]", "", 0, nil)
	debugView := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetWordWrap(true).
		SetChangedFunc(func() {
			indices := tv.bufferList.List.FindItems("debug", "", true, false)
			if len(indices) != 0 {
				tv.bufferList.List.SetItemText(indices[0], "[pink]debug **[white]", "")
			}
		})
	tv.pages.AddPage("page-debug", debugView, true, false)
	tv.buffers["debug"] = debugView
	return debugView
}
