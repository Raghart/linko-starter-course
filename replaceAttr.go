package main

import (
	"fmt"
	"log/slog"
)

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "error" {
		err, ok := a.Value.Any().(error)
		if !ok {
			return a
		}
		return slog.String("error", fmt.Sprintf("%+v", err))
	}
	return a
}
