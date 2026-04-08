package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type closeFunc func() error

func closeHelper(function func() error) {
	if err := function(); err != nil {
		log.Printf("There was a problem while trying to flush the helper: %v", err)
	}
}

func initializeLogger(logFile string) (*log.Logger, closeFunc, error) {
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("error while trying to initiliaze a logger:", err)
		}
		bufferedFile := bufio.NewWriterSize(file, 8192)

		multiWriter := io.MultiWriter(os.Stderr, bufferedFile)
		return log.New(multiWriter, "", log.LstdFlags), func() error {
			defer file.Close()
			return bufferedFile.Flush()
		}, nil
	}
	return log.New(os.Stderr, "", log.LstdFlags), func() error { return nil }, nil
}
