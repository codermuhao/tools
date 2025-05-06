// Package xjson xjson
package xjson

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	jsoniter "github.com/json-iterator/go"
)

var (
	// marshalOptions is a configurable JSON format marshaller.
	marshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
		UseEnumNumbers:  true,
	}
	// unmarshalOptions is a configurable JSON format parser.
	unmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// initialisms golint定义：https://github.com/golang/lint/blob/master/lint.go#L770
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var initialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
}

// Marshal json marshal
func Marshal(v interface{}) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		return marshalOptions.Marshal(m)
	}
	return json.Marshal(v)
}

// Unmarshal json unmarshal
func Unmarshal(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	for rv := rv; rv.Kind() == reflect.Ptr; {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if m, ok := v.(proto.Message); ok {
		return fallbackUnmarshal(unmarshalOptions.Unmarshal(data, m), data, v)
	} else if m, ok := reflect.Indirect(rv).Interface().(proto.Message); ok {
		return fallbackUnmarshal(unmarshalOptions.Unmarshal(data, m), data, v)
	}
	return json.Unmarshal(data, v)
}

// fallbackUnmarshal 为了兼容类似struct定义为int，而收到的是string的情况
func fallbackUnmarshal(err error, data []byte, v interface{}) error {
	if err == nil {
		return nil
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           v,
		TagName:          "json",
		WeaklyTypedInput: true,
		MatchName: func(mapKey, fieldName string) bool {
			try1 := strings.EqualFold(mapKey, fieldName)
			if try1 {
				return true
			}
			return initialismsEqual(mapKey, strings.ToUpper(fieldName[:1])+fieldName[1:])
		},
	})
	if err != nil {
		return err
	}
	return decoder.Decode(m)
}

func initialismsEqual(mapKey, fieldName string) bool {
	var fun = func(key string) string {
		runes := []rune(key)
		w, i := 0, 0 // index of start of word, scan
		for i+1 <= len(runes) {
			eow := false // whether we hit the end of a word
			if i+1 == len(runes) {
				eow = true
			} else if runes[i+1] == '_' {
				// underscore; shift the remainder forward over any run of underscores
				eow = true
				n := 1
				for i+n+1 < len(runes) && runes[i+n+1] == '_' {
					n++
				}

				// Leave at most one underscore if the underscore is between two digits
				if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
					n--
				}

				copy(runes[i+1:], runes[i+n+1:])
				runes = runes[:len(runes)-n]
			} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
				// lower->non-lower
				eow = true
			}
			i++
			if !eow {
				continue
			}

			// [w,i) is a word.
			word := string(runes[w:i])
			if u := strings.ToUpper(word); initialisms[u] {
				// Keep consistent case, which is lowercase only at the start.
				if w == 0 && unicode.IsLower(runes[w]) {
					u = strings.ToLower(u)
				}
				// All the common initialisms are ASCII,
				// so we can replace the bytes exactly.
				copy(runes[w:], []rune(u))
			} else if w > 0 && strings.ToLower(word) == word {
				// already all lowercase, and not the first word, so uppercase the first character.
				runes[w] = unicode.ToUpper(runes[w])
			}
			w = i
		}
		return string(runes)
	}
	return fun(mapKey) == fun(fieldName)
}
