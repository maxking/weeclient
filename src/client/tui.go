// Terminal based relay client for Weechat.
package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/maxking/weeclient/src/weechat"
	"github.com/rivo/tview"
)

type TerminalView struct {
	app        *tview.Application
	grid       *tview.Grid
	sendchan   chan string
	bufferList *BufferListWidget
	pages      *tview.Pages
	buffers    map[string]*tview.TextView
}

// Event handler when something in a buffer widget changes.
func (tv *TerminalView) SetCurrentBuffer(index int, mainText, secondaryText string, shortcut rune) {
	// special handlinge for the debug buffer with and without unread count.
	if mainText == "[red]debug[white]" || mainText == "[pink]debug **[white]" {
		tv.pages.SwitchToPage("page-debug")
		// remove the unread aspect.
		tv.bufferList.List.SetItemText(index, "[red]debug[white]", "")
		return
	}
	// Handle weechat buffers.
	buf := tv.bufferList.getByFullName(mainText)
	if buf != nil {
		// For the buffer widget, set the right number of lines.
		if bufView, ok := tv.buffers[buf.FullName]; ok {
			//			tv.app.QueueUpdate(func() {
			bufView.SetText(buf.TitleStr(true) + buf.GetLines(true))
			tv.bufferList.List.SetItemText(index, fmt.Sprintf("%v", buf.FullName), "")
			// Then, switch to the page that is embedding the above buffer widget.
			tv.pages.SwitchToPage(fmt.Sprintf("page-%v", buf.FullName))
			// Send command to load nicklist of the buffer if there
			// is no nicklist in it and it is a channel not person (# check)
			if buf.FullName != "debug" && buf.NickList.GetItemCount() == 0 && strings.Contains(buf.FullName, "#") {
				tv.sendchan <- fmt.Sprintf("(nicklist) nicklist %v\n", buf.FullName)
			}
			// })
		} else {
			tv.Debug(
				fmt.Sprintf(
					"Failed to find the buffer in tv.buffers %v buffername %v\n",
					mainText, buf.FullName))
		}

	} else {
		tv.Debug(fmt.Sprintf("Failed to find the buffer %v\n", mainText))
	}
}

func (tv *TerminalView) FocusBuffer(index int, mainText, SecondaryTest string, shortcut rune) {
	tv.app.SetFocus(tv.pages)
}

func TviewStart(
	weechan chan *weechat.WeechatMessage, sendchan chan string) {
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
