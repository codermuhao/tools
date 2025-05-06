package test

import (
	"strings"
	"testing"

	"github.com/codermuhao/tools/xjson"
	"google.golang.org/protobuf/proto"
)

func TestXMarshalPB(t *testing.T) {
	tests := []struct {
		input  proto.Message
		expect string
	}{
		{
			input:  &BigInt{},
			expect: `{"bigint_fixed64":0,"bigint_int64":0,"bigint_sfixed64":0,"bigint_sint64":0,"bigint_uint64":0}`,
		},
		{
			input:  &BigInt{BigintInt64: 1, BigintFixed64: 2, BigintSfixed64: 3, BigintSint64: 4, BigintUint64: 5},
			expect: `{"bigint_fixed64":2,"bigint_int64":1,"bigint_sfixed64":3,"bigint_sint64":4,"bigint_uint64":5}`,
		},
	}
	for _, v := range tests {
		data, err := xjson.XMarshalPB(v.input)
		if err != nil {
			t.Errorf("marshal(%#v): %s", v.input, err)
		}
		if got, want := string(data), v.expect; got != want {
			if strings.Contains(want, "\n") {
				t.Errorf("marshal(%#v):\nHAVE:\n%s\nWANT:\n%s", v.input, got, want)
			} else {
				t.Errorf("marshal(%#v):\nhave %#q\nwant %#q", v.input, got, want)
			}
		}
	}
}
