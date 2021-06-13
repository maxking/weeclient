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
	return o.Value.(uint32) == 0
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

// All the information about a new line.
type WeechatLine struct {
	// Path of the buffer.
	Buffer      string
	Date        string
	DatePrinted string
	Displayed   bool
	NotifyLevel int
	Highlight   bool
	Tags        []string
	Prefix      string
	Message     string
}

func (wb WeechatBuffer) AddLine(message string) {
	wb.Lines = append(wb.Lines, message)
}

// Interface for handler that handles various events.
type HandleWeechatMessage interface {
	HandleListBuffers(map[string]*WeechatBuffer)

	HandleListLines()

	HandleNickList(*WeechatMessage)

	HandleLineAdded(*WeechatLine)

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
		buflist := make(map[string]*WeechatBuffer, len(bufffers.Value))

		for _, each := range bufffers.Value {
			buf := &WeechatBuffer{
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

	case "_buffer_line_added", "listlines":
		for _, each := range msg.Object.Value.(WeechatHdaValue).Value {
			line := WeechatLine{
				Buffer:  each["buffer"].as_string(),
				Message: each["message"].as_string(),
				Date:    each["date"].as_string(),
				// DatePrinted: each["date_printed"].as_string(),
				// Displayed:   each["displayed"].as_bool(),
				// NotifyLevel: each["notify_level"].as_int(),
				// Highlight: each["highlight"].as_bool(),
				Prefix: each["prefix"].as_string(),
			}
			handler.HandleLineAdded(&line)
		}
		handler.HandleListLines()
		// add the lines to a buffer.
	case "nicklist":
		// handle list of nicks.
		handler.HandleNickList(msg)
	default:
		handler.Default(msg)
	}
	return nil
}
