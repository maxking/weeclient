package client

// client package primarily holds all the code for a shell UI.
import (
	"fmt"
	"log"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/maxking/weeclient/src/weechat"
)

type WeechatTerminalUI struct {
	AllBuffers    map[string]*weechat.WeechatBuffer
	CurrentBuffer string
	grid          *ui.Grid
	buflist       *widgets.List
	buffer        *widgets.Paragraph
}

func (w *WeechatTerminalUI) Init() {
	w.grid = ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	w.grid.SetRect(0, 0, termWidth, termHeight)
	w.buflist = w.getBufferListWidget()
	w.buffer = w.getBufferWidget(w.CurrentBuffer)
	w.grid.Set(
		ui.NewCol(1.0/4, w.buflist),
		ui.NewCol(3.0/4, w.buffer),
	)
}

func (w *WeechatTerminalUI) Draw() {
	ui.Render(w.grid)
}

func (w *WeechatTerminalUI) UpdateBufferlist(buffers map[string]*weechat.WeechatBuffer) {
	var buflist []string
	for _, buf := range buffers {
		buflist = append(buflist, buf.FullName)
	}
	w.buflist.Rows = buflist
	w.AllBuffers = buffers
}

func (w *WeechatTerminalUI) getBufferListWidget() *widgets.List {
	ls := widgets.NewList()
	var buffers []string
	for _, buf := range w.AllBuffers {
		buffers = append(buffers, fmt.Sprintf("%v. %v", buf.Number, buf.FullName))
	}
	ls.Rows = buffers
	ls.Title = "Buffers"
	return ls
}

func (w *WeechatTerminalUI) getBufferWidget(bufname string) *widgets.Paragraph {
	p := widgets.NewParagraph()
	if len(w.AllBuffers) == 0 {
		p.Text = "Hello World!"
	} else {
		var lines []string
		for _, buf := range w.AllBuffers {
			lines = append(lines, buf.Lines...)
		}
		// buf := w.AllBuffers[bufname]
		p.Text = strings.Join(lines, "\n")
		// p.Title = buf.FullName
		p.Title = "All Buffers"
	}
	// p.SetRect(0, 0, 25, 5)
	return p
}

func (w *WeechatTerminalUI) AddLine(buffer, message string) {
	w.AllBuffers[buffer].AddLine(message)
}

func (w *WeechatTerminalUI) SetCurrentBuffer(buffer string) {
	if len(w.AllBuffers) == 0 || buffer == "" {
		w.buffer.Text = "Hello World!"
	} else {
		buf := w.AllBuffers[buffer]
		w.buffer.Text = strings.Join(buf.Lines, "\n")
		w.buffer.Title = buf.FullName
	}
}

func (w *WeechatTerminalUI) GetBufferShortname(bufferAddr string) string {
	if bufferAddr == "" || len(w.AllBuffers) == 0 {
		return "Invalid Input"
	}
	if buf, ok := w.AllBuffers[bufferAddr]; !ok {
		return "Not Found"
	} else {
		return buf.ShortName
	}
}

func Start(weechan chan *weechat.WeechatMessage) {
	// Primary entrypoint to start the terminal ui and the event loop
	// that manages everything related to the terminal rendering and
	// reacting to the updates.
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	allBuffers := make(map[string]*weechat.WeechatBuffer)

	terminalUi := WeechatTerminalUI{
		AllBuffers:    allBuffers,
		CurrentBuffer: "",
	}
	messageHandler := &TerminalMessageHandler{&terminalUi}
	// instantiate the whole ui.
	terminalUi.Init()
	terminalUi.Draw()

	// Get a channel for polling for events.
	uiEvents := ui.PollEvents()

	// Run an infinite loop that polls for events form terminal or
	// from weechat and take actions accordingly.
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				// Quit.
				return
			case "<Resize>":
				// Re-draw the entire page again upon resize.
				terminalUi.Draw()
			}
		case weemsg := <-weechan:
			// Handle a new message.
			weechat.HandleMessage(weemsg, messageHandler)
		}
	}
}

type TerminalMessageHandler struct {
	terminalUI *WeechatTerminalUI
}

func (mh *TerminalMessageHandler) HandleListBuffers(buflist map[string]*weechat.WeechatBuffer) {
	mh.terminalUI.UpdateBufferlist(buflist)
	mh.terminalUI.Draw()

}

func (mh *TerminalMessageHandler) HandleListLines() {
	mh.terminalUI.Draw()
}

func (mh *TerminalMessageHandler) HandleNickList() {

}

func (mh *TerminalMessageHandler) HandleLineAdded(line *weechat.WeechatLine) {
	mh.terminalUI.AddLine(line.Buffer, line.Message)
	mh.terminalUI.buffer.Text = fmt.Sprintf("%v\n%v : %v",
		mh.terminalUI.buffer.Text, mh.terminalUI.GetBufferShortname(line.Buffer), line.Message)

}

func (mh *TerminalMessageHandler) Default(msg *weechat.WeechatMessage) {
	// fmt.Println(msg)
}
