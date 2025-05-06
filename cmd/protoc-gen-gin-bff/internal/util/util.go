// Package util util
package util

import (
	"strings"
)

// FirstLower 首字母小写
func FirstLower(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToLower(name[:1]) + name[1:]
}

// FormatType 格式化rpc方法的请求和返回参数
func FormatType(name string) string {
	if len(name) == 0 {
		return ""
	}
	//return name
	switch name {
	case ".google.protobuf.Empty":
		return "emptypb.Empty"
	default:
		//return name
		parts := strings.Split(name, ".")
		return parts[len(parts)-1]
	}
}

// VariableName 生成变量名
func VariableName(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(s, "/", ""),
		"-",
		"",
	)
}
