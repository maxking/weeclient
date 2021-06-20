// Handle colors in Weechat strings.
package color

import (
	"fmt"
	"regexp"
)

const (
	TitleColor   = "red"
	DefaultColor = "white"
	TimeColor    = "green"
	MsgColor     = "white"
	ChatColor    = "grey"
	ChanColor    = "blue"
	NickColor    = "pink"
	UnreadColor  = "purple"
	// Color for messages that have particular sub-strings,
	// like join and leave
	JoinColor       = "grey"
	LeaveColor      = "grey"
	NickChangeColor = "grey"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

// Parse out colors of the form [color]text[color]
// from the string if it exists.
func RemoveColor(with_color string) string {
	re := regexp.MustCompile(`\[.*\](.*)\[.*\]`)
	return re.ReplaceAllString(with_color, "$1")
}

var (
	// Mostly copied verbatim from qweechat.
	// https://github.com/weechat/qweechat/blob/master/qweechat/weechat/color.py#L25
	ColorsRegex = `[*!/_]*`
	ColorsStd   = fmt.Sprintf(`(?:%s\d{2})`, ColorsRegex)
	ColorsExt   = fmt.Sprintf(`(?:@%v\d{5})`, ColorsRegex)
	ColorsAny   = fmt.Sprintf(`(?:%s|%s)`, ColorsStd, ColorsExt)
	ColorsRe    = fmt.Sprintf(
		`(\x19(?:\d{2}|F%v|B\d{2}|B@\d{5}|E|\\*%v(,%v)?|@\d{5}|b.|\x1C))|\x1A.|\x1B.|\x1C`,
		ColorsAny, ColorsAny, ColorsAny)
)

// Replace the weechat colors parsed using regex and use the replaceFund to
// find substituations.
func ReplaceWeechatColors(with_color string, replacefunc func(string) string) string {
	re := regexp.MustCompile(ColorsRe)
	return re.ReplaceAllStringFunc(with_color, replacefunc)
}

// When fully implemented, will replace weechat colors with colors that
// tview understands. Currently, not in a functional state.
func Colorize(in string) string {
	switch in[0] {
	case '\x19':
		switch in[1] {
		case 'b':
			return ""
		case '\x1C':
			return "\x01(Fr)\x01(Br)"
		default:
			return ""
		}
	default:
		return ""
	}
}
