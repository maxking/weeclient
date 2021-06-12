package weechat

/// Core weechat object. This represents a parsed Core object type.
/// It doesn't currently capture all the data in a single object
/// and kinda uses only a single untyped Value fields. In future,
/// try to figure out something better here that alows optional
/// fields for specific objects.
type WeechatObject struct {
	ObjType string
	Value   interface{}
}

/// Represents a single message from weechat.
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
	Lines     []string
	ShortName string
	FullName  string
	Number    uint32
	Title     string
	LocalVars map[WeechatObject]WeechatObject
	Path      string
}

func (wb WeechatBuffer) AddLine(message string) {
	wb.Lines = append(wb.Lines, message)
}

// Interface for handler that handles various events.
type HandleWeechatMessage interface {
	HandleListBuffers(map[string]WeechatBuffer)

	HandleListLines()

	HandleNickList()

	HandleLineAdded(string, string)

	Default(*WeechatMessage)
}

// Parse the message into More useful data structures that can be used by higher
// level UI functions. It expects an interface which handles parsed structured
// output.
func HandleMessage(msg *WeechatMessage, handler HandleWeechatMessage) error {
	switch msg.Msgid {
	case "listbuffers":
		// parse out the list of buffers which are Hda objects.
		bufffers := msg.Object.Value.(WeechatHdaValue)
		buflist := make(map[string]WeechatBuffer, len(bufffers.Value))

		for _, each := range bufffers.Value {
			buf := WeechatBuffer{
				ShortName: each["short_name"].Value.(string),
				FullName:  each["full_name"].Value.(string),
				Title:     each["title"].Value.(string),
				Number:    each["number"].Value.(uint32),
				LocalVars: each["local_variables"].Value.(map[WeechatObject]WeechatObject),
				Lines:     []string{""},
				// this is essentially a list of strings, pointers,
				// the first pointer of which is the buffer' pointer.
				Path: each["__path"].Value.([]string)[1],
			}
			buflist[buf.Path] = buf
		}

		handler.HandleListBuffers(buflist)

	case "listlines":
		// handle list of lines in each buffer.
		for _, each := range msg.Object.Value.(WeechatHdaValue).Value {
			handler.HandleLineAdded(each["buffer"].Value.(string), each["message"].Value.(string))

			// fmt.Printf("---\nbuffer = %v message = %v \n---\n", each["buffer"], each["message"])
		}
		handler.HandleListLines()
	case "_buffer_line_added":
		for _, each := range msg.Object.Value.(WeechatHdaValue).Value {
			handler.HandleLineAdded(each["buffer"].Value.(string), each["message"].Value.(string))
		}
		handler.HandleListLines()
		// add the lines to a buffer.
	// case "nicklist":
	// handle list of nicks.
	// fmt.Printf("%v", msg.Value)
	default:
		handler.Default(msg)
	}
	return nil
}
