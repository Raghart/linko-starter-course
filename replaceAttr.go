package main

import (
	"errors"
	"fmt"
	"log/slog"

	pkgerr "github.com/pkg/errors"
)

type stackTracer interface {
	error
	StackTrace() pkgerr.StackTrace
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "error" {
		err, ok := a.Value.Any().(error)
		if !ok {
			return a
		}
		if stackErr, ok := errors.AsType[stackTracer](err); ok {
			return slog.GroupAttrs("error", slog.Attr{
				Key:   "message",
				Value: slog.StringValue(stackErr.Error()),
			}, slog.Attr{
				Key:   "stack_trace",
				Value: slog.StringValue(fmt.Sprintf("%+v", stackErr.StackTrace())),
			})
		}
		return slog.String("error", fmt.Sprintf("%+v", err))
	}
	return a
}
