# xerrors
[TOC]

## 背景

在一个提供服务的程序中，自身的error处理以及如何给调用者清晰的错误码是非常重要的功能。 为了统一我们的错误码方案及错误处理流程，封装了xerrors库。

## 功能特性

1. 完全兼容pkg/errors
2. 支持自定义业务错误，方便调用者判断
3. 自定义错误可携带元数据
4. 自定义错误支持`Continue`，该标识用于说明发生错误后业务能否正常进行

## 更新日志

参见 [CHANGELOG](CHANGELOG.md)

## 安装

```shell
go get github.com/codermuhao/tools/xerrors
```

## 使用说明

```go
package main
// 暂略
```

## 注意事项

暂无

## 相关仓库

* [pkg errors][1] — Package errors provides simple error handling primitives.
* [kratos][2] - a microservice-oriented governance framework implemented by golang

## 维护者

* @yitttang
* @xaviexiang

[1]: https://github.com/pkg/errors
[2]: https://github.com/go-kratos/kratos
