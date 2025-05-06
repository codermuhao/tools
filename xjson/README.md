# 项目名称
[TOC]

## 背景

对json序列号和烦序列号进行封装，支持普通json操作和pb message的json操作，减轻开发者使用负担

## 功能特性

1. 自动判断使用通用json还是protobuf结构的json操作
2. 由于历史项目原因，支持`XMarshalPB`操作，该操作有设计缺陷，禁止随便使用
3. 增加proto unmarshal的降级处理（兼容例如struct定义为int，而收到的是string）

## 更新日志

参见 [CHANGELOG](CHANGELOG.md)

## 安装

```shell
go get github.com/codermuhao/tools/xjson
```

## 使用说明

```go
package main

import (
	"fmt"
	"github.com/codermuhao/tools/xjson"
)

type Message struct {
	Field1 string     `json:"a"`
	Field2 string     `json:"b"`
	Field3 string     `json:"c"`
	Field4 int64      `json:"d"`
}

func main() {
	m, err := xjson.Marshal(Message{
		Field1: "f1",
		Field2: "f2",
		Field3: "f3",
		Field4: 1233,
	})
	if err != nil {
		panic(err)
    }
	fmt.Println(string(m))
	
	var m1, m2 Message
	s1 := `{"a":"f1", "b":"f2", "c":"f3", "d":1233}`
	if err := xjson.Unmarshal([]byte(s1), &m1);err != nil {
		panic(err)
    }

    // 自动降级处理处理，string => int64
	s2 := `{"a":"f1", "b":"f2", "c":"f3", "d":"1233"}`
	if err := xjson.Unmarshal([]byte(s2), &m2);err != nil {
		panic(err)
	}
	
}
```

## 注意事项

* 谨慎使用`XMarshalPB`，禁止乱用

## 相关仓库

* [protobuf (github.com)][1] — Go support for Protocol Buffers
* [json-iterator][2] - A high-performance 100% compatible drop-in replacement of "encoding/json"
* [protobuf (github.com)][3] - Go support for Protocol Buffers

## 维护者

* @yitttang 


[1]: https://github.com/golang/protobuf
[2]: https://github.com/RichardLitt/standard-readme
[3]: https://google.golang.org/protobuf