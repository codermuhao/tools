package test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codermuhao/tools/xjson"
)

func TestProto_JSON_Marshal(t *testing.T) {
	tests := []struct {
		input  interface{}
		expect string
	}{
		{
			input:  &Outer{},
			expect: `{"outer_string":"", "inner":null, "status":0}`,
		},
		{
			input:  &Outer{OuterString: "outer_string", Status: Status_Status_Failure},
			expect: `{"outer_string":"outer_string", "inner":null, "status":2}`,
		},
		{
			input: &Outer{OuterString: "outer_string", Status: Status_Status_Success, Inner: &Inner{
				InnerString:        "inner_string",
				InnerInt:           12,
				InnerBool:          false,
				InnerRepeatedFloat: []float32{1.1, 2.2, 3.3},
			}},
			expect: `{"outer_string":"outer_string", "inner":{"inner_string":"inner_string", "inner_int":12, "inner_bool":false, "inner_repeated_float":[1.1, 2.2, 3.3]}, "status":1}`,
		},
	}
	for _, v := range tests {
		data, err := xjson.Marshal(v.input)
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

func TestProto_JSON_Unmarshal(t *testing.T) {
	p1, p2 := &Outer{}, &Outer{}
	_ = p1
	_ = p2
	tests := []struct {
		input  string
		expect interface{}
	}{
		{
			input:  `{"outer_string":"", "inner":null, "status":0}`,
			expect: &Outer{},
		},
		{
			input:  `{"outer_string":"outer_string","status":2}`,
			expect: &p1,
		},
		{
			input:  `{"outer_string":"outer_string","inner":{"inner_string":"inner_string","inner_int":12,"inner_bool":true,"inner_repeated_float":[1.1,2.2,3.3]},"status":1}`,
			expect: &p2,
		},
	}
	for _, v := range tests {
		want := []byte(v.input)
		err := xjson.Unmarshal(want, &v.expect)
		if err != nil {
			t.Errorf("marshal(%#v): %s", v.input, err)
		}
		got, err := xjson.Marshal(v.expect)
		if err != nil {
			t.Errorf("marshal(%#v): %s", v.input, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("marshal(%#v):\nhave %#q\nwant %#q", v.input, got, want)
		}
	}
}
