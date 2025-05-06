// Package util util
package util

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var enCases = cases.Title(language.AmericanEnglish, cases.NoLower)

// Case2Camel 下划线转为驼峰
func Case2Camel(name string) string {
	if !strings.Contains(name, "_") {
		upperName := strings.ToUpper(name)
		if upperName == name {
			name = strings.ToLower(name)
		}
		return enCases.String(name)
	}
	name = strings.Replace(strings.ToLower(name), "_", " ", -1)
	return strings.Replace(enCases.String(name), " ", "", -1)
}
