package client

import (
	"fmt"

	"github.com/maxking/weeclient/src/weechat"
	"github.com/rivo/tview"
)

type BufferListWidget struct {
	List    *tview.List
	Buffers map[string]*weechat.WeechatBuffer
}

func (bw *BufferListWidget) getByFullName(fullname string) *weechat.WeechatBuffer {
	for _, buf := range bw.Buffers {
		if buf.FullName == fullname {
			return buf
		}
	}
	return nil
}

func NewBufferListWidget(buflist map[string]*weechat.WeechatBuffer) *BufferListWidget {
	widget := &BufferListWidget{
		List:    tview.NewList().ShowSecondaryText(false),
		Buffers: buflist,
	}
	return widget
}

func (w *BufferListWidget) AddBuffer(buffer *weechat.WeechatBuffer) {

}

type TerminalView struct {
	app  *tview.Application
	grid *tview.Grid

	bufferList *BufferListWidget
	buffer     *tview.TextView
}

// Event handlers.
func (tv *TerminalView) SetCurrentBuffer(index int, mainText, secondaryText string, shortcut rune) {
	buf := tv.bufferList.getByFullName(mainText)
	if buf != nil {
		tv.buffer.SetText(fmt.Sprintf("index: %v primary: %v secondary: %v\n----\n %v\n%v",
			index, mainText, secondaryText, buf.FullName, buf.Title))
	} else {
		tv.buffer.SetText(fmt.Sprintf("Buffer not found! \nindex: %v primary: %v secondary: %v\n%v\n=====\n",
			index, mainText, secondaryText, tv.bufferList))
	}
}

// Methods for Weechat message Handler.
func (tv *TerminalView) HandleListBuffers(buflist map[string]*weechat.WeechatBuffer) {
	for ptr, buf := range buflist {
		tv.bufferList.List.AddItem(buf.FullName, "", 'u', nil)
		tv.bufferList.Buffers[ptr] = buf
	}
}

func (tv *TerminalView) HandleListLines() {

}

func (tv *TerminalView) HandleNickList() {

}

func (tv *TerminalView) HandleLineAdded(buffer, message string) {

}

func (tv *TerminalView) Default(msg *weechat.WeechatMessage) {

}

func TviewStart(weechan chan *weechat.WeechatMessage) {
	app := tview.NewApplication()

	textView := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Sample")

	bufffers := make(map[string]*weechat.WeechatBuffer)
	buflist := NewBufferListWidget(bufffers)

	grid := tview.NewGrid().
		SetColumns(-1, -4).
		SetBorders(true).
		AddItem(buflist.List, 0, 0, 1, 1, 0, 0, true).
		AddItem(textView, 0, 1, 1, 1, 0, 0, false)

	view := &TerminalView{app: app, grid: grid, bufferList: buflist, buffer: textView}
	view.bufferList.List.SetChangedFunc(view.SetCurrentBuffer)

	// Read from the weechat incoming queue and enquee for handling.
	go func() {
		for msg := range weechan {
			weechat.HandleMessage(msg, view)
		}
	}()

	if err := view.app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}
