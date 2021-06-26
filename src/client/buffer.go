package client

import (
	"strings"

	"github.com/maxking/weeclient/src/color"
	"github.com/maxking/weeclient/src/weechat"

	"github.com/rivo/tview"
)

type Buffer struct {
	*weechat.WeechatBuffer
	Chat     *tview.TextView
	Users    *tview.List
	Input    *tview.InputField
	NickList *tview.List
}

type BufferListWidget struct {
	List    *tview.List
	Buffers map[string]*Buffer
}

func (bw *BufferListWidget) getByFullName(fullname string) *Buffer {
	name := color.RemoveColor(fullname)
	for _, buf := range bw.Buffers {
		if buf.FullName == name {
			return buf
		}
	}
	// panic(fmt.Sprintf("fullname: %v name: %v  -- failed to find buffer by fullname", fullname, name))
	return nil
}

// Create a new buffer list widget.
func NewBufferListWidget(buflist map[string]*Buffer) *BufferListWidget {
	widget := &BufferListWidget{
		List:    tview.NewList().ShowSecondaryText(false),
		Buffers: buflist,
	}
	return widget
}

// Add a new buffer to the buffer list widget.
// Sorting order is:
// core.weechat
// irc.server....
// irc.<server>.#channel
// irc.<server>.nick
func (w *BufferListWidget) AddBuffer(buffer string) {
	var index int
	if buffer == "core.weechat" {
		index = 0
	} else if strings.Contains(buffer, "irc.server") {
		index = 1
	} else if strings.Contains(buffer, "#") {
		index = 2
	} else {
		index = -1
	}
	w.List.InsertItem(index, buffer, "", 0, nil)
}
