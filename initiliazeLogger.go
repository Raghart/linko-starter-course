package main

import (
	"io"
	"log"
	"os"
)

func initializeLogger(logFile string) *log.Logger {
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("error while trying to initiliaze a logger:", err)
		}

		multiWriter := io.MultiWriter(os.Stderr, file)
		return log.New(multiWriter, "", log.LstdFlags)
	}
	return log.New(os.Stderr, "", log.LstdFlags)
}
