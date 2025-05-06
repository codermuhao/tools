// Package xerrors xerrors
package xerrors

import (
	"fmt"

	"github.com/codermuhao/tools/xjson"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"

	microerrors "github.com/asim/go-micro/v3/errors"
)

var unknown = "UNKNOWN_ERROR"

// ReasonError 带reason和元信息的error结构
type ReasonError struct {
	Msg      string                 `json:"msg"`
	Reason   string                 `json:"reason"`
	Continue bool                   `json:"continue"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewReasonError returns an error object for the reason, message.
func NewReasonError(reason string, message string) *ReasonError {
	return &ReasonError{
		Reason:   reason,
		Msg:      message,
		Metadata: make(map[string]interface{}),
	}
}

// NewReasonErrorf NewReasonError(reason fmt.Sprintf(format, args...))
func NewReasonErrorf(reason, format string, args ...interface{}) *ReasonError {
	return NewReasonError(reason, fmt.Sprintf(format, args...))
}

// Parse try to convert an error to *Error.
// It supports wrapped errors.
func Parse(err error) *ReasonError {
	if err == nil {
		return nil
	}
	if se := new(ReasonError); errors.As(err, &se) {
		return se
	}
	gs, ok := status.FromError(err)
	if ok {
		for _, detail := range gs.Details() {
			switch d := detail.(type) {
			case *errdetails.ErrorInfo:
				md := make(map[string]interface{}, len(d.Metadata))
				for k, v := range d.Metadata {
					md[k] = v
				}
				return NewReasonError(d.Reason, gs.Message()).WithMetadata(md)
			}
		}
		return NewReasonError(unknown, err.Error())
	}
	switch err.(type) {
	case *microerrors.Error:
		se := new(ReasonError)
		if err := xjson.Unmarshal([]byte(err.(*microerrors.Error).Detail), &se); err == nil {
			return se
		}
	default:
		se := new(ReasonError)
		if err := xjson.Unmarshal([]byte(err.Error()), &se); err == nil {
			return se
		}
	}
	return NewReasonError(unknown, err.Error())
}

// WithMetadata with an MD formed by the mapping of key, value.
func (e *ReasonError) WithMetadata(md map[string]interface{}) *ReasonError {
	for k, v := range md {
		e.Metadata[k] = v
	}
	return e
}

// WithContinue 设置continue字段为true表示此类错误是业务正常流程错
// 该标记可以作为是否记录日志等操作的依据
func (e *ReasonError) WithContinue() *ReasonError {
	e.Continue = true
	return e
}

// Continue returns the 'continue' flag.
func Continue(err error) bool {
	if err == nil {
		return true
	}
	if se := Parse(err); err != nil {
		return se.Continue
	}
	return false
}

// Error implements error interface
func (e *ReasonError) Error() string {
	data, err := xjson.Marshal(e)
	if err != nil {
		data, _ = xjson.Marshal(NewReasonError(unknown, err.Error()))
	}
	return string(data)
}

// Is matches each error in the chain with the target value.
func (e *ReasonError) Is(err error) bool {
	if se := new(ReasonError); errors.As(err, &se) {
		return se.Reason == e.Reason
	}
	se := new(ReasonError)
	if err := xjson.Unmarshal([]byte(err.Error()), &se); err == nil {
		return se.Reason == e.Reason
	}
	return false
}
