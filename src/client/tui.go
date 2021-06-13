package client

import (
	"fmt"
	"os"
	"strings"

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
	pages      *tview.Pages
	buffers    map[string]*tview.TextView
}

// Event handlers.
func (tv *TerminalView) SetCurrentBuffer(index int, mainText, secondaryText string, shortcut rune) {
	buf := tv.bufferList.getByFullName(mainText)
	if buf != nil {
		// For the buffer widget, set the right number of lines.
		if bufView, ok := tv.buffers[buf.FullName]; ok {
			bufView.SetText((fmt.Sprintf("[%v] %v\n-----\n%v", buf.FullName, buf.Title, strings.Join(buf.Lines, "\n"))))
		}
		// Then, switch to the page that is embedding the above buffer widget.
		tv.pages.SwitchToPage(fmt.Sprintf("page-%v", mainText))
		// tv.app.SetFocus(tv.pages)
	}
}

// Methods for Weechat message Handler.
func (tv *TerminalView) HandleListBuffers(buflist map[string]*weechat.WeechatBuffer) {
	for ptr, buf := range buflist {
		tv.bufferList.List.AddItem(buf.FullName, "", 'u', nil)
		tv.bufferList.Buffers[ptr] = buf
		bufferView := tview.NewTextView().
			SetTextAlign(tview.AlignLeft).
			SetWordWrap(true).
			SetDynamicColors(true).
			SetText(fmt.Sprintf("[%v] %v\n-----\n%v", buf.FullName, buf.Title, strings.Join(buf.Lines, "\n")))

		bufferView.SetTitle(buf.FullName)
		bufferView.SetChangedFunc(func() { tv.app.Draw() })

		tv.pages.AddPage(fmt.Sprintf("page-%v", buf.FullName),
			bufferView,
			true,
			// Only the core.weechat buffer is visible at first.
			buf.FullName == "core.weechat",
		)

		tv.buffers[buf.FullName] = bufferView

	}
}

func (tv *TerminalView) HandleListLines() {

}

func (tv *TerminalView) HandleNickList() {

}

func (tv *TerminalView) HandleLineAdded(line *weechat.WeechatLine) {
	buf := tv.bufferList.Buffers[line.Buffer]
	buf.Lines = append(buf.Lines, line.Message)
	// Also, add the message to the current view.
	if bufView, ok := tv.buffers[buf.FullName]; ok {
		bufView.Write([]byte(fmt.Sprintf("\n[%v] <%v>: %v", line.Date, line.Prefix, line.Message)))
	}
}

func (tv *TerminalView) Default(msg *weechat.WeechatMessage) {

}

func TviewStart(weechan chan *weechat.WeechatMessage) {
	app := tview.NewApplication()
	bufffers := make(map[string]*weechat.WeechatBuffer)
	buflist := NewBufferListWidget(bufffers)
	bufferspage := tview.NewPages()
	bufferViews := make(map[string]*tview.TextView, 100)

	grid := tview.NewGrid().
		SetColumns(-1, -4).
		SetBorders(true).
		AddItem(buflist.List, 0, 0, 1, 1, 0, 0, true).
		AddItem(bufferspage, 0, 1, 1, 1, 0, 0, false)

	view := &TerminalView{app: app, grid: grid, bufferList: buflist, pages: bufferspage, buffers: bufferViews}
	view.bufferList.List.SetChangedFunc(view.SetCurrentBuffer)

	// Read from the weechat incoming queue and enquee for handling.
	go func() {
		for msg := range weechan {
			weechat.HandleMessage(msg, view)
		}
	}()

	if err := view.app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		// panic(err)
		fmt.Println(fmt.Errorf("Error from the application: %v", err))
		os.Exit(1)
	}
}
