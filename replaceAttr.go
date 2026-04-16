package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"boot.dev/linko/internal/linkoerr"
	pkgerr "github.com/pkg/errors"
)

type stackTracer interface {
	error
	StackTrace() pkgerr.StackTrace
}

type multiError interface {
	error
	Unwrap() []error
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	var sensitiveKeys = []string{
		"password",
		"key",
		"apiKey",
		"secret",
		"pin",
		"creditcardno",
		"user",
	}

	if a.Key == "longURL" {
		urlParsed, _ := url.Parse(a.Value.String())
		completeURL := urlParsed.String()
		pass, _ := urlParsed.User.Password()
		parsedPass := strings.Replace(completeURL, pass, "[REDACTED]", -1)
		return slog.String("long_url", parsedPass)
	}

	if slices.Contains(sensitiveKeys, a.Key) {
		return slog.String(a.Key, "[REDACTED]")
	}

	if a.Key == "error" {
		err, ok := a.Value.Any().(error)
		if !ok {
			return a
		}

		if me, ok := errors.AsType[multiError](err); ok {
			var errAttrs []slog.Attr
			for i, e := range me.Unwrap() {
				errAttrs = append(errAttrs,
					slog.GroupAttrs(fmt.Sprintf("error_%d", i+1), errorAttrs(e)...))
			}
			return slog.GroupAttrs("errors", errAttrs...)
		}

		return slog.GroupAttrs("error", errorAttrs(err)...)
	}
	return a
}

func errorAttrs(err error) []slog.Attr {
	attr := []slog.Attr{
		{Key: "message", Value: slog.StringValue(err.Error())},
	}
	attr = append(attr, linkoerr.Attrs(err)...)
	if stackErr, ok := errors.AsType[stackTracer](err); ok {
		attr = append(attr, slog.Attr{
			Key:   "stack_trace",
			Value: slog.StringValue(fmt.Sprintf("%+v", stackErr.StackTrace())),
		})
	}
	return attr
}
