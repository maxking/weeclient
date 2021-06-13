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
		List:    tview.NewList(),
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
		tv.buffer.SetText(fmt.Sprintf("index: %v primary: %v secondary: %v\n\n\n----\n %v\n%v",
			index, mainText, secondaryText, buf.FullName, buf.Title))
	} else {
		tv.buffer.SetText(fmt.Sprintf("Buffer not found! \nindex: %v primary: %v secondary: %v\n%v\n\n=====\n",
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
	// list := tview.NewList().
	// 	AddItem("List item 1", "Some explanatory text", 'a', nil).
	// 	AddItem("List item 2", "Some explanatory text", 'b', nil).
	// 	AddItem("List item 3", "Some explanatory text", 'c', nil).
	// 	AddItem("List item 4", "Some explanatory text", 'd', nil)

	// list.AddItem("Quit", "Press to exit", 'q', func() {
	// 	app.Stop()
	// })
	// list.AddItem("Add One", "Press to add a new entry", 'n', func() {
	// 	list.AddItem(fmt.Sprintf("List item %v", list.GetItemCount()+1),
	// 		"Newly added item", 'o', nil)
	// })

	// Read from the weechat incoming queue.
	go func() {
		for msg := range weechan {
			weechat.HandleMessage(msg, view)
		}
	}()

	if err := view.app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}
