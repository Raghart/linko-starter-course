package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logCtx := &LogContext{}
			r = r.WithContext(context.WithValue(r.Context(), logContextKey, logCtx))

			spyReader := &spyReadClose{ReadCloser: r.Body}
			r.Body = spyReader

			spyWriter := &spyResponseWriter{ResponseWriter: w}
			next.ServeHTTP(spyWriter, r)

			methodSlogSlice := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("client_ip", redactIP(r.RemoteAddr)),
				slog.Duration("duration", time.Since(start)),
				slog.Int("request_body_bytes", spyReader.bytesRead),
				slog.Int("response_status", spyWriter.statusCode),
				slog.Int("response_body_bytes", spyWriter.bytesWritten),
			}

			if logCtx.Username != "" {
				methodSlogSlice = append(methodSlogSlice, slog.String("user", logCtx.Username))
			}

			if err, ok := logCtx.Error.(error); ok {
				methodSlogSlice = append(methodSlogSlice, slog.GroupAttrs("error",
					errorAttrs(err)...))
			}

			if headerID := r.Header.Get("X-Request-ID"); strings.TrimSpace(headerID) != "" {
				methodSlogSlice = append(methodSlogSlice, slog.String("request_id", headerID))
			}

			logger.Info("Served request",
				methodSlogSlice...,
			)
		})
	}
}

func requestHeader() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerID := r.Header.Get("X-Request-ID")
			if strings.TrimSpace(headerID) == "" {
				headerID = rand.Text()
			}
			w.Header().Set("X-Request-ID", headerID)
			next.ServeHTTP(w, r)
		})
	}
}

func redactIP(adress string) string {
	host, _, err := net.SplitHostPort(adress)
	if err != nil {
		return adress
	}

	parsedIP := net.ParseIP(host)
	if parsedIP.DefaultMask() == nil {
		return host
	}

	ip4 := parsedIP.To4()

	return fmt.Sprintf("%v.%v.%v.x", ip4[0], ip4[1], ip4[2])
}
