/*
Weechat allow decoding a message from weechat relay.

Parser can be initialized as `weechat.Protocol` object that has
a `Decode()` method which parses a full message. There is a private
method to parse every objtype and `parseObject` takes in the objType
and data returning the parsed object. Currently, `parseObject` can
parse all the Core weechat object types.

There are a few more public methods, `ParseString` and `ParseInt`
which provide public methods to parse the string and uint32 types.

## Protocol Parsing

`Decode` method returns a `WeechatMessage` object which includes
various methods including a list of `WeechatObjects` and `Msgid`.
They are the most used fields, but there is also the size and
compression related information captured in the `WeechatMessage`
object instances.

`WeechatObject`s are currently naive in implementation and uses
a single Type with `ObjType` and `Value` parameters that captures
the Core type of the obj and _any_ value type. You have to know
the type of the Value or check it explicitly by looking at `ObjType`.

HDA is the most common data type. We use `map[string]*WeechatObject`
as the Value type, but in order to also capture other details like
hpath, we created a wrapper around the value and Hda object's values
are of `WeechatHdaValue` type. For now, it only has a `Value` and
`Hpath` attributes, but in future we might add other fields if
needed or even remove it.

All the parsing related code lives in protocol.go

## Message Handling

Once a raw response has been parsed, `weechat.HandleMessage` is
used to parse useful information from the Core types and return
more useful objects by looking at `WeechatMessage.Msgid`.

Msgid for commands can be specified in the commands and the response
will have those Msgids. Weechat also has some specific Msgids
that it uses to signify the event type.

Currently, there are three commands that weeclient sends on start,
apart from the authentication with custom Msgids. The rest are
essentially the default Msgids.

- `(listbuffers) hdata buffer:gui_buffers(*) number,full_name,short_name,type,nicklist,title,local_variables,`
- `(listlines) hdata buffer:gui_buffers(*)/own_lines/last_line(-%(lines)d)/data date,displayed,prefix,message,buffer`
- `(nicklist) nicklist`

### Currently supported events:

`weechat.HandleMessage` currently only parses *some* Msgid types
and then uses a catch-all handler for the rest of the types. It
is up-to the callee to handle the rest of the types.

`weechat.HandleMessage` expects a handler of the type `HandleWeechatMessage`
interface, which essentially is meant to be an interface that handles all
the [Core event types in Weechat][1].

Currently handled events include:

- *listbuffers* - This is a custom command which you can find above
- *listlines* - This is also a custom command.
- =*nicklist* - This is a custom command too. Although, `HandleNickList()` is
  called on the msg object with no additional object types or parsing.

Weechat events that are handled:
- *_buffer_opened*: When a new buffer is opened. Handler is same as "listbuffers"
  and `HandleListBuffers()` is called on the handler.
- *_buffer_line_added*: When a new line is added in any of the buffer. Handler is
  same as "listlines" and `HandleLineAdded` is called.

[1]: https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#message_identifier




*/
package weechat
