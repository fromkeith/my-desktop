package main

import (
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

func SetupJsonEncoding() {
	extra.SetNamingStrategy(func(name string) string {
		if name == "" {
			return name
		}
		r, size := utf8.DecodeRuneInString(name)
		return string(unicode.ToLower(r)) + name[size:]
	})
	RegisterTimeFormat()
}
func RegisterTimeFormat() {
	jsoniter.RegisterTypeEncoder("time.Time", &timeAsRfc3339Nano{})
	jsoniter.RegisterTypeDecoder("time.Time", &timeAsRfc3339Nano{})
}

type timeAsRfc3339Nano struct {
}

func (codec *timeAsRfc3339Nano) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	var str = iter.ReadString()
	t, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		return
	}
	*((*time.Time)(ptr)) = t
}

func (codec *timeAsRfc3339Nano) IsEmpty(ptr unsafe.Pointer) bool {
	return ((*time.Time)(ptr)) == nil
}
func (codec *timeAsRfc3339Nano) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteString(ts.Format(time.RFC3339Nano))
}
