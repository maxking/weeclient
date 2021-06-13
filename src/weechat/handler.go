package weechat

import (
	"strconv"
	"time"
)

// Interface for handler that handles various events.
type HandleWeechatMessage interface {
	HandleListBuffers(map[string]*WeechatBuffer)

	HandleNickList(*WeechatMessage)

	HandleLineAdded(*WeechatLine)

	Default(*WeechatMessage)
}

// Parse the message into More useful data structures that can be used by higher
// level UI functions. It expects an interface which handles parsed structured
// output.
func HandleMessage(msg *WeechatMessage, handler HandleWeechatMessage) error {
	switch msg.Msgid {
	case "listbuffers", "_buffer_opened":
		// parse out the list of buffers which are Hda objects.
		bufffers := msg.Object.Value.(WeechatHdaValue)
		buflist := make(map[string]*WeechatBuffer, len(bufffers.Value))

		for _, each := range bufffers.Value {
			buf := &WeechatBuffer{
				ShortName: each["short_name"].Value.(string),
				FullName:  each["full_name"].Value.(string),
				Title:     each["title"].Value.(string),
				Number:    each["number"].Value.(uint32),
				LocalVars: each["local_variables"].Value.(map[WeechatObject]WeechatObject),
				Lines:     make([]*WeechatLine, 0),
				// this is essentially a list of strings, pointers,
				// the first pointer of which is the buffer' pointer.
				Path: each["__path"].Value.([]string)[1],
			}
			buflist[buf.Path] = buf
		}

		handler.HandleListBuffers(buflist)

	case "_buffer_line_added", "listlines":
		for _, each := range msg.Object.Value.(WeechatHdaValue).Value {
			secs, _ := strconv.ParseInt(each["date"].as_string(), 10, 64)
			utime := time.Unix(secs, 0)
			line := WeechatLine{
				Buffer:  each["buffer"].as_string(),
				Message: each["message"].as_string(),
				Date:    utime,
				// DatePrinted: each["date_printed"].as_string(),
				Displayed: each["displayed"].as_bool(),
				// NotifyLevel: each["notify_level"].as_int(),
				Highlight: each["highlight"].as_bool(),
				Prefix:    each["prefix"].as_string(),
			}
			handler.HandleLineAdded(&line)
		}
		// add the lines to a buffer.
	case "nicklist":
		// handle list of nicks.
		handler.HandleNickList(msg)

	default:
		handler.Default(msg)
	}
	return nil
}
