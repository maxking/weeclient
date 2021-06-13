package weechat

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Protocol represents a parser for Weechat Relay Protocol and essentially
// parses messages from here:
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#messages
type Protocol struct {
}

// All the Core Weechat objecct types.
const (
	OBJ_INT = "int"
	OBJ_CHR = "chr"
	OBJ_LON = "lon"
	OBJ_STR = "str"
	OBJ_BUF = "buf"
	OBJ_PTR = "ptr"
	OBJ_TIM = "tim"
	OBJ_HTB = "htb"
	OBJ_HDA = "hda"
	OBJ_INF = "inf"
	OBJ_INL = "inl"
	OBJ_ARR = "arr"
)

// This is the primary Pubic method to decode a single Weechat Message. It
// support zlib decompression of the compressed message.
func (p *Protocol) Decode(data []byte) (*WeechatMessage, error) {
	var objType, msgid string
	msglen, compressed, msgBody, _ := p.parseInitial(data)
	if compressed {
		var out bytes.Buffer
		in := bytes.NewReader(msgBody)
		r, err := zlib.NewReader(in)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to decompress zlib compressed data")
		}
		io.Copy(&out, r)
		r.Close()
		msgBody = out.Bytes()
	}
	msgid, msgBody = p.ParseString(msgBody)
	objType, msgBody = p.parseType(msgBody)
	obj, _ := p.parseObject(objType, msgBody)
	// fmt.Printf("Total size: %v\nCompression: %v\nId: %v\nType: %v\n======\n",
	// 	msglen, compressed, msgid, objType)
	return &WeechatMessage{
		Size:             int(msglen),
		Compressed:       compressed,
		SizeUncompressed: len(msgBody),
		Msgid:            msgid,
		Type:             objType,
		Object:           obj,
	}, nil
}

// Parse the objects as defined in the list here:
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#objects
func (p *Protocol) parseObject(
	objType string, data []byte) (WeechatObject, []byte) {

	switch objType {
	case OBJ_HDA:
		return p.parseHda(data)
	case OBJ_TIM:
		return p.parseTime(data)
	case OBJ_CHR:
		return p.parseChar(data)
	case OBJ_STR:
		return p.parseStr(data)
	case OBJ_INT:
		return p.parseInt(data)
	case OBJ_HTB:
		return p.parseHashTable(data)
	case OBJ_ARR:
		return p.parseArray(data)
	case OBJ_INF:
		return p.parseInfo(data)
	case OBJ_INL:
		return p.parseInfoListContent(data)
	case OBJ_LON:
		return p.parseLongInt(data)
	case OBJ_PTR:
		return p.parsePointer(data)
	default:
		panic(fmt.Errorf("parsing of type %v is not implemented yet", objType))
	}
}

// Parse the length of the message body.
func (p *Protocol) parseInitial(data []byte) (uint32, bool, []byte, []byte) {
	len, _ := p.ParseLen(data)
	compressed := bytes.Equal(data[4:5], []byte{01})
	msgBody := data[5:len]
	return len, compressed, msgBody, data[len:]
}

// Parse a single string.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_string
func (p *Protocol) ParseString(data []byte) (string, []byte) {
	var len uint32
	len, data = p.ParseLen(data)
	// 2147483647 is the value for int_max in 4byte singed integer.
	if len <= 0 || len >= 2147483647 {
		return "", data
	}
	// fmt.Printf("Length of string: %v value: %v\n", len, str)
	return string(data[:len]), data[len:]
}

// Parse a single string.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_string
func (p *Protocol) parseStr(data []byte) (WeechatObject, []byte) {
	var strval string
	strval, data = p.ParseString(data)
	return WeechatObject{OBJ_STR, strval}, data
}

// Parse a 3 letter object type, usually one of:
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#objects
// There is currently no check built to handle object types we don't know
// about and it will essentially end up panic in parseObject if an unknown
// type is passed to it.
func (p *Protocol) parseType(data []byte) (string, []byte) {
	return string(data[:3]), data[3:]
}

// Parse the length of a string/message/anything. Just a
// proxy for parseInt.
func (p *Protocol) ParseLen(data []byte) (uint32, []byte) {
	var val WeechatObject
	val, data = p.parseInt(data)
	return val.Value.(uint32), data
}

// Parse 4 byte signed Integer.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_integer
func (p *Protocol) parseInt(data []byte) (WeechatObject, []byte) {
	if len(data) < 4 {
		return WeechatObject{OBJ_INT, 0}, data
	}
	len := binary.BigEndian.Uint32(data[:4])
	// fmt.Printf("Parsed int of value: %v\n", len)
	return WeechatObject{OBJ_INT, len}, data[4:]
}

// Parse time, which is essentially a string with a 1 byte length
// in the start.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_time
func (p *Protocol) parseTime(data []byte) (WeechatObject, []byte) {
	var pointer WeechatObject
	pointer, data = p.parsePointer(data)
	pointer.ObjType = OBJ_TIM
	return pointer, data
}

// Parse a single character of length 1 byte.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_char
func (p *Protocol) parseChar(data []byte) (WeechatObject, []byte) {
	return WeechatObject{OBJ_CHR, string(data[0])}, data[1:]
}

// Parse a hash table datatype. It starts with two Type (3byte) (key type, value type)
// and then the count (4 byte integer) and then count number of key value pairs.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_hashtable
func (p *Protocol) parseHashTable(data []byte) (WeechatObject, []byte) {
	var key_type, value_type string
	var count uint32
	key_type, data = p.parseType(data)
	value_type, data = p.parseType(data)
	count, data = p.ParseLen(data)

	var key, value WeechatObject
	hashtable := make(map[WeechatObject]WeechatObject, count)

	// fmt.Printf("Key types: %v, value types: %v\n", key_type, value_type)
	for i := 0; i < int(count); i++ {
		key, data = p.parseObject(key_type, data)
		value, data = p.parseObject(value_type, data)
		hashtable[key] = value
	}
	// fmt.Println("=====Finished parsing of hash table.")
	return WeechatObject{OBJ_HTB, hashtable}, data
}

// Parse an array. Objects of a single type.
// 3 bytes Type, 4 byte count (integer) and then count number of objects.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_array
func (p *Protocol) parseArray(data []byte) (WeechatObject, []byte) {
	var objType string
	var count uint32
	var value WeechatObject
	// parse the type of the objects in the array.
	objType, data = p.parseType(data)

	count, data = p.ParseLen(data)
	arr := make([]WeechatObject, count)

	for i := 0; i < int(count); i++ {
		value, data = p.parseObject(objType, data)
		arr = append(arr, value)
	}

	return WeechatObject{OBJ_ARR, arr}, data
}

// Parse Infolist, which is a list of key-value pairs with key type string and
// arbitrary value type.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_infolist
func (p *Protocol) parseInfoListContent(data []byte) (WeechatObject, []byte) {
	var count uint32
	var objType, key string
	var value WeechatObject
	// parse name
	_, data = p.parseStr(data)
	// parse count
	count, data = p.ParseLen(data)
	// parse count numer of items.

	infolist := make(WeechatDict, count)

	for i := 0; i < int(count); i++ {
		// parse name.
		key, data = p.ParseString(data)
		// parse type.
		objType, data = p.parseType(data)
		// parse the value.
		value, data = p.parseObject(objType, data)
		infolist[key] = value

	}
	return WeechatObject{OBJ_INL, infolist}, data
}

// Parse Info, which is essentially a key-value pair of type string.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_info
func (p *Protocol) parseInfo(data []byte) (WeechatObject, []byte) {
	info := make(map[string]string)
	var key, value string
	// key
	key, data = p.ParseString(data)
	// value
	value, data = p.ParseString(data)
	info[key] = value

	return WeechatObject{OBJ_INF, info}, data
}

// Parse Long integer, which is essentially parsed like a Pointer,
// single byt length and then string.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_long_integer
func (p *Protocol) parseLongInt(data []byte) (WeechatObject, []byte) {
	var pointer WeechatObject
	pointer, data = p.parsePointer(data)
	pointer.ObjType = OBJ_LON
	return pointer, data
}

// Parse hdata. This is the most complex data structure to be
// parsed so just read the explanation in docs.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_hdata
func (p *Protocol) parseHda(data []byte) (WeechatObject, []byte) {
	// fmt.Println("========= Parsing object: hda ========")
	var hpath, keys string
	var remaining []byte
	var count uint32
	// Parse the hpath.
	hpath, remaining = p.ParseString(data)
	// parse the keys.
	keys, remaining = p.ParseString(remaining)
	// parse the count.
	count, remaining = p.ParseLen(remaining)

	// fmt.Printf("hpath: %v\nkeys: %v\ncount: %v\n", hpath, keys, count)

	pointerCount := len(strings.Split(hpath, "/"))

	// Parse objects based on keys "count" number of times.
	objects := strings.Split(keys, ",")
	// fmt.Printf("hapth: %v\n", hpath)

	// Hdata is modeled as list of dictionaries where the key type is
	// string and the value type is defined in the "keys" section, but
	// is essentially consistent across the list.
	var hda []WeechatDict
	var objVal WeechatObject
	var pointers []string

	for j := 0; j < int(count); j++ {

		// Parse out the pointers, there is no use of them that I
		// can think of. Perhaps, in future, we can use it somewhere.
		pointers, remaining = p.parsePointers(pointerCount, remaining)

		hdamap := make(WeechatDict, len(objects)+1)
		// sorta bad pattern here to jam in the pointers into the hda
		// map so that it can be used later.
		hdamap["__path"] = WeechatObject{"__path", pointers}
		for _, obj := range objects {
			s := strings.Split(obj, ":")
			objName, objtype := s[0], s[1]
			objVal, remaining = p.parseObject(objtype, remaining)
			// fmt.Printf("%v:%v. Parsed %v of type %v value\n",
			// 	j, i, objname, objtype)
			hdamap[objName] = objVal
		}
		// Now add it to the hda list.
		hda = append(hda, hdamap)
	}
	// fmt.Printf("hpath:%v\nkeys:%v\ncount:%v\npointerCount: %v\nvalues:%v\n",
	// hpath, keys, count, pointerCount, values)
	// fmt.Println("========= Finished parsing hda ========")

	// Wrap the values into a specific values object that includes the Hpath
	return WeechatObject{OBJ_HDA, WeechatHdaValue{Value: hda, Hpath: hpath}}, remaining
}

// Parse count number of pointers. Not an actual datatype.
func (p *Protocol) parsePointers(count int, data []byte) ([]string, []byte) {
	pointers := make([]string, count)
	var pointer WeechatObject
	for i := 0; i < count; i++ {
		pointer, data = p.parsePointer(data)
		pointers = append(pointers, pointer.Value.(string))
	}
	// fmt.Printf("Parsed %v pointers\n", count)
	return pointers, data
}

// Parse a single byte integer, used to denote length of
// pointer or long integer. Not an actual datatype.
func (p *Protocol) parseSmallint(data []byte) int {
	return int(data[0])
}

// Parse a single pointer. Single byte length and then length
// long string.
// https://weechat.org/files/doc/stable/weechat_relay_protocol.en.html#object_pointer
func (p *Protocol) parsePointer(data []byte) (WeechatObject, []byte) {
	len := p.parseSmallint(data)
	if len == 0 {
		return WeechatObject{OBJ_PTR, ""}, data[2:]
	}
	pointer := string(data[1 : len+1])
	return WeechatObject{OBJ_PTR, pointer}, data[len+1:]
}
