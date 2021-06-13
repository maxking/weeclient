// Terminal based relay client for Weechat.
package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/maxking/weeclient/src/color"
	"github.com/maxking/weeclient/src/weechat"
	"github.com/rivo/tview"
)

const (
	titleColor   = "red"
	defaultColor = "white"
	timeColor    = "green"
	msgColor     = "yellow"
	chatColor    = "grey"
)

type TerminalView struct {
	app        *tview.Application
	grid       *tview.Grid
	sendchan   chan *weechat.WeechatSendMessage
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
			bufView.SetText((fmt.Sprintf("[%v] [%v] %v\n-----\n%v [%v]",
				msgColor, buf.FullName, buf.Title,
				color.StripWeechatColors(strings.Join(buf.Lines, "\n"), color.Colorize),
				defaultColor)))
		}
		// Then, switch to the page that is embedding the above buffer widget.
		tv.pages.SwitchToPage(fmt.Sprintf("page-%v", mainText))
	}
}

func (tv *TerminalView) FocusBuffer(index int, mainText, SecondaryTest string, shortcut rune) {
	tv.app.SetFocus(tv.pages)
}

func TviewStart(
	weechan chan *weechat.WeechatMessage, sendchan chan *weechat.WeechatSendMessage) {
	app := tview.NewApplication()
	bufffers := make(map[string]*Buffer)
	buflist := NewBufferListWidget(bufffers)
	bufferspage := tview.NewPages()
	bufferViews := make(map[string]*tview.TextView, 100)

	grid := tview.NewGrid().
		SetColumns(-1, -4).
		SetBorders(true).
		AddItem(buflist.List, 0, 0, 1, 1, 0, 0, true).
		AddItem(bufferspage, 0, 1, 1, 1, 0, 0, false)

	// Create a terminalview object which holds all the state
	// for the current state of the terminal.
	view := &TerminalView{
		app:        app,
		grid:       grid,
		bufferList: buflist,
		pages:      bufferspage,
		buffers:    bufferViews,
		sendchan:   sendchan}
	view.bufferList.List.SetChangedFunc(view.SetCurrentBuffer)
	view.bufferList.List.SetSelectedFunc(view.FocusBuffer)

	// Set keybindings to move the focus back to buffer list.
	view.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlB {
			view.app.SetFocus(view.bufferList.List)
		}
		return event
	})

	// Read from the weechat incoming queue and enquee for handling.
	go func() {
		for msg := range weechan {
			weechat.HandleMessage(msg, view)
		}
	}()

	if err := view.app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		// panic(err)
		fmt.Println(fmt.Errorf("error from the application: %v", err))
		os.Exit(1)
	}
}
