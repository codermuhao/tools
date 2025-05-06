package xjson

import (
	"encoding/base64"

	"google.golang.org/protobuf/proto"
	pref "google.golang.org/protobuf/reflect/protoreflect"
)

// XMarshalPB protojson特殊版本，强烈建议不要使用！！！
// 仅为了老代码保留，主要改动是不把64位整数转为string，但实际上这是有一定风险的，具体请见：
// https://stackoverflow.com/questions/53911502
func XMarshalPB(pb proto.Message) ([]byte, error) {
	v := itrMessage(pb.ProtoReflect())
	j, err := json.Marshal(v)

	return j, err
}

func itrMessage(pm pref.Message) interface{} {
	m := make(map[string]interface{})
	fds := pm.Descriptor().Fields()
	for i := 0; i < fds.Len(); {
		fd := fds.Get(i)

		if od := fd.ContainingOneof(); od != nil {
			fd = pm.WhichOneof(od)
			i += od.Fields().Len()
			if fd == nil {
				continue // unpopulated oneofs are not affected by EmitUnpopulated
			}
		} else {
			i++
		}

		val := pm.Get(fd)
		if !pm.Has(fd) {
			isProto2Scalar := fd.Syntax() == pref.Proto2 && fd.Default().IsValid()
			isSingularMessage := fd.Cardinality() != pref.Repeated && fd.Message() != nil
			if isProto2Scalar || isSingularMessage {
				// Use invalid value to emit null.
				val = pref.Value{}
			}
		}

		name := string(fd.Name())
		// Use type name for group field name.
		if fd.Kind() == pref.GroupKind {
			name = string(fd.Message().Name())
		}

		switch {
		case fd.IsList():
			m[name] = itrList(val.List(), fd)
		case fd.IsMap():
			m[name] = itrMap(val.Map(), fd)
		default:
			m[name] = itrSingular(val, fd)
		}

	}

	return m
}

func itrList(list pref.List, fd pref.FieldDescriptor) interface{} {
	l := make([]interface{}, 0)
	for i := 0; i < list.Len(); i++ {
		item := list.Get(i)
		l = append(l, itrSingular(item, fd))
	}
	return l
}

// mapEntry kv实体
type mapEntry struct {
	key   pref.MapKey
	value pref.Value
}

func itrMap(mmap pref.Map, fd pref.FieldDescriptor) interface{} {
	m := make(map[string]interface{})

	// Get a sorted list based on keyType first.
	entries := make([]mapEntry, 0, mmap.Len())
	mmap.Range(func(key pref.MapKey, val pref.Value) bool {
		entries = append(entries, mapEntry{key: key, value: val})
		return true
	})

	// Write out sorted list.
	for _, entry := range entries {
		m[entry.key.String()] = itrSingular(entry.value, fd.MapValue())
	}
	return m
}

func itrSingular(val pref.Value, fd pref.FieldDescriptor) interface{} {

	if !val.IsValid() {
		m := make(map[string]interface{})
		return m
	}

	var v interface{}
	switch kind := fd.Kind(); kind {
	case pref.BoolKind:
		v = val.Bool()

	case pref.StringKind:
		v = val.String()

	case pref.Int32Kind, pref.Sint32Kind, pref.Sfixed32Kind,
		pref.Int64Kind, pref.Sint64Kind, pref.Sfixed64Kind:
		v = val.Int()

	case pref.Uint32Kind, pref.Fixed32Kind,
		pref.Uint64Kind, pref.Fixed64Kind:
		v = val.Uint()

	case pref.FloatKind:
		v = val.Float()

	case pref.DoubleKind:
		v = val.Float()

	case pref.BytesKind:
		v = base64.StdEncoding.EncodeToString(val.Bytes())

	case pref.EnumKind:
		v = int64(val.Enum())

	case pref.MessageKind, pref.GroupKind:
		v = itrMessage(val.Message())
	default:
	}

	return v
}
