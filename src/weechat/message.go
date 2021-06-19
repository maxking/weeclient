package weechat

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/maxking/weeclient/src/color"
)

// Core weechat object. This represents a parsed Core object type.
// It doesn't currently capture all the data in a single object
// and kinda uses only a single untyped Value fields. In future,
// try to figure out something better here that alows optional
// fields for specific objects.
type WeechatObject struct {
	ObjType string
	Value   interface{}
}

// Coerce the value to string. Only handles string types for now.
func (o WeechatObject) as_string() string {
	return o.Value.(string)

	// switch o.ObjType {
	// case OBJ_STR:
	// 	return o.Value.(string)
	// default:
	// 	return "Invalid string coercion of Weechat obj"
	// }
}

func (o WeechatObject) as_int() uint32 {
	return o.Value.(uint32)
}

func (o WeechatObject) as_bool() bool {
	return o.Value == '0'
}

// Object representing information needed to be sent.
type WeechatSendMessage struct {
	Message string
	Buffer  string
}

// Represents a single message from weechat.
type WeechatMessage struct {
	// Size of the message when recieved including the length (4bytes).
	Size int

	// Size of the message after (optional) decompressing.
	SizeUncompressed int

	// Was zlib compressed used in the message body?
	Compressed bool

	// Uncompressed content of the message. If it wasn't compressed
	// this has the originl body of the message minus the length.
	Uncompressed []byte

	// optional message-id of the message.
	Msgid string

	// Object type.
	Type string

	// List of Weechat objects returned from the message.
	Object WeechatObject
}

type WeechatDict map[string]WeechatObject

type WeechatHdaValue struct {
	Value []WeechatDict
	Hpath string
}

type WeechatBuffer struct {
	Lines     []*WeechatLine
	ShortName string
	FullName  string
	Number    int32
	Title     string
	LocalVars map[WeechatObject]WeechatObject
	Path      string
}

// Get the Title of the Buffer with color if asked for.
func (b *WeechatBuffer) TitleStr(shouldColor bool) string {
	if shouldColor {
		return fmt.Sprintf("[%v][%v][%v] %v[%v]\n---\n",
			color.ChanColor, b.FullName, color.TitleColor, b.Title, color.DefaultColor)
	}
	return fmt.Sprintf("[%v] %v\n---\n", b.FullName, b.Title)
}

func (b *WeechatBuffer) GetLines(shouldColor bool) string {
	var lines []string
	for _, line := range b.Lines {
		lines = append(lines, color.ReplaceWeechatColors(line.ToString(shouldColor), color.Colorize))
	}
	return strings.Join(lines, "\n")
}

// All the information about a new line.
type WeechatLine struct {
	// Path of the buffer.
	Buffer      string
	Date        time.Time
	DatePrinted time.Time
	Displayed   bool
	NotifyLevel int
	Highlight   bool
	Tags        []string
	Prefix      string
	Message     string
}

// Return the string representation of the line to be printed in the
// ui. Use optional coloring.
func (l *WeechatLine) ToString(shouldColor bool) string {
	if shouldColor {
		// WHen using [color] for coloring the output, we want to make sure
		// the actual text within square braces isn't lost trying to color
		// the output. To do that, we need to escape it by replacing `]` by
		// `[]` resulting in something like `[hello[]` to print `[hello]`
		// https://pkg.go.dev/github.com/rivo/tview@v0.0.0-20210608105643-d4fb0348227b?utm_source=gopls#hdr-Colors
		re := regexp.MustCompile(`\]`)
		msg := string(re.ReplaceAll([]byte(l.Message), []byte("[]")))
		return fmt.Sprintf("[%v][%v] [%v] %v: [%v] %v[%v]",
			color.TimeColor, l.Date.Format("15:05"),
			color.NickColor, l.Prefix,
			color.MsgColor, color.ReplaceWeechatColors(msg, color.Colorize),
			color.DefaultColor)
	}
	return fmt.Sprintf("[%v:%v] %v: %v",
		l.Date.Hour(), l.Date.Minute(),
		l.Prefix,
		// Replace colors with just nothing.
		color.ReplaceWeechatColors(l.Message, func(s string) string { return "" }))

}
