package client

import (
	"github.com/maxking/weeclient/src/color"
	"github.com/maxking/weeclient/src/weechat"

	"github.com/rivo/tview"
)

type Buffer struct {
	*weechat.WeechatBuffer
	Chat  *tview.TextView
	Users *tview.List
	Input *tview.InputField
}

type BufferListWidget struct {
	List    *tview.List
	Buffers map[string]*Buffer
}

func (bw *BufferListWidget) getByFullName(fullname string) *weechat.WeechatBuffer {
	name := color.RemoveColor(fullname)
	for _, buf := range bw.Buffers {
		if buf.FullName == name {
			return buf.WeechatBuffer
		}
	}
	// panic(fmt.Sprintf("fullname: %v name: %v  -- failed to find buffer by fullname", fullname, name))
	return nil
}

func NewBufferListWidget(buflist map[string]*Buffer) *BufferListWidget {
	widget := &BufferListWidget{
		List:    tview.NewList().ShowSecondaryText(false),
		Buffers: buflist,
	}
	return widget
}

func (w *BufferListWidget) AddBuffer(buffer *weechat.WeechatBuffer) {

}
