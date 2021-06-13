package color

import (
	"fmt"
	"regexp"
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
	ColorsRegex = `[*!/_]*`
	ColorsStd   = fmt.Sprintf(`(?:%s\d{2})`, ColorsRegex)
	ColorsExt   = fmt.Sprintf(`(?:@%v\d{5})`, ColorsRegex)
	ColorsAny   = fmt.Sprintf(`(?:%s|%s)`, ColorsStd, ColorsExt)
	ColorsRe    = fmt.Sprintf(`(\x19(?:\d{2}|F%v|B\d{2}|B@\d{5}|E|\\*%v(,%v)?|@\d{5}|b.|\x1C))|\x1A.|\x1B.|\x1C`, ColorsAny, ColorsAny, ColorsAny)
)

func StripWeechatColors(with_color string) string {
	re := regexp.MustCompile(ColorsRe)
	return re.ReplaceAllString(with_color, "")
}
